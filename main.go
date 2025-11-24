package main

import (
    "bufio"
    "fmt"
    "net"
    "os"
    "os/exec"
    "time"
	"golang.org/x/sys/windows"
	"crypto/sha256"
	"strings"
    "operclite-1/gui"
    "operclite-1/mod_gui"
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/app"
    "strconv"
)

func main() {
    //Logica de interconexion
    var operclite fyne.App = app.New()
    clearScreen()
    if (len(os.Args) < 4) {
        fmt.Println("Parametros: operclite.exe <ip> <port> <time interval for reports>")
        return
    }
    ipRemota := strings.TrimSpace(os.Args[1])
    ptoRemoto := strings.TrimSpace(os.Args[2])
    setReportInterval := strings.TrimSpace(os.Args[3])
    dirTCP, _ := net.ResolveTCPAddr("tcp4", ipRemota+":"+ptoRemoto)
    serverSocket, _ := net.DialTCP("tcp4", nil, dirTCP)
	//Recibir mensaje por defecto de el servidor para saber si la conexión es correcta o no
	if recInvalidKey(serverSocket) == "Unauthorized\r\n" { //Conexion invalida
        gui.ErrorWindow(operclite)
		serverSocket.Close()
		os.Exit(0)
	}
    fmt.Println("\n--------------------------------------")
    fmt.Println("*        Operclite Client Login        *")
    fmt.Println("----------------------------------------")
    sendCredentialMsg(serverSocket, "Time: " + setReportInterval + "\n")
    fmt.Println("Intervalo enviado")
	if !loginProcess(serverSocket, operclite) {
        return
    }
    clearScreen()
    mainWin, gui_access, report_access := gui.StartMainWindow(operclite)
    quitC := make(chan bool)
	//Logica principal de comunicación
    go recMsgMain(serverSocket, quitC, gui_access, report_access)
    gui_access.SendButton.OnTapped = func() {
        envMsgMain(serverSocket, quitC, gui_access)
    }
    go func() {
        <-quitC
        serverSocket.Close()
        fyne.Do(func() {
            gui.ByeWindow(operclite)
            (*mainWin).Close()
        })
    }()
    (*mainWin).SetIcon(resourceIconPng)
    (*mainWin).ShowAndRun() 
}

/*Proceso de login, se recibe por parametro el socket y la app con el fin de invocar una ventana de error en caso
de que el cliente quede bloqueado  (esto sucedera si el el tiempo exponencial llega a superar los 3 minutos)*/
func loginProcess(socket net.Conn, app fyne.App) bool {
    for {
        recDefaultMsg(socket) //Recibe mensaje de ingresar usuario
		fmt.Print("Ingrese el usuario: ")
		reader := bufio.NewReader(os.Stdin)
        usercli, _ := reader.ReadString('\n')
        sendCredentialMsg(socket, usercli) //Envia la credencial

		recDefaultMsg(socket) //Recibe mensaje de ingresar contraseña
		fmt.Print("Ingrese la contraseña: ")
		hPasswd, err := readPassword()
	    if err != nil {
		    fmt.Println("Error:", err)
		    os.Exit(1)
	    }
		sendCredentialMsg(socket, hPasswd) //Envia la credencial
		tempKey := recInvalidKey(socket)
        if strings.HasPrefix(tempKey, "Freeze: ") {
            counter := strings.TrimSpace(strings.TrimPrefix(tempKey, "Freeze: "))
            //Transformamos a entero
            counterInt, err := strconv.Atoi(counter)
            if err != nil {
                fmt.Println("Error al transformar en entero: ", err)
                return false
            }
            fmt.Println("Muchos intentos por parte del cliente, poner en espera durante ", counterInt)
            time.Sleep(time.Duration(counterInt) * time.Second)
            continue
        }
        switch(tempKey) {
        case "Retry\r\n": //Ingreso mal las credenciales, vuelve a intentarlo hasta que el servidor suelte un Freeze o un Locked
            continue
        case "Login OK\r\n": //Rompe el proceso de login para seguir con la intercomunicacion normal
            fmt.Println("Login correcto, iniciando comunicacion...")
            return true
        case "Locked\r\n": //Bloqueado
            clearScreen()
            socket.Close()
            gui.ErrorWindow(app)
            return false
        default:
            return false
        }
    }
}

//----------Leer las contraseñas de manera correcta para Windows-------------//
func readPassword() (string, error) {
	h := windows.Handle(os.Stdin.Fd())
	var oldMode uint32
	if err := windows.GetConsoleMode(h, &oldMode); err != nil {
		return "", fmt.Errorf("no se pudo obtener el modo de consola: %v", err)
	}

	newMode := oldMode &^ windows.ENABLE_ECHO_INPUT
	if err := windows.SetConsoleMode(h, newMode); err != nil {
		return "", fmt.Errorf("no se pudo desactivar el eco: %v", err)
	}

	reader := bufio.NewReader(os.Stdin)
	passwd, err := reader.ReadString('\n')
	_ = windows.SetConsoleMode(h, oldMode)
	fmt.Println()
	if err != nil {
		return "", fmt.Errorf("error al leer la contraseña: %v", err)
	}
	passwd = strings.TrimSpace(passwd)
	hashTemp := sha256.Sum256([]byte(passwd))
	hPasswd := fmt.Sprintf("%x", hashTemp)
	return hPasswd + "\n", nil
}

//--------Funciones de comunicacion de manera concurrente (Incluye la GUI)---------//

func envMsgMain(socket net.Conn, channel chan bool, gui *gui.GUIComponents) {
    msgEnv := gui.UserEntry.Text
    if strings.TrimSpace(msgEnv) == "" {
        return
    }
    env := bufio.NewWriter(socket)
    env.WriteString(msgEnv + "\n")
    env.Flush()
    gui.AddMessage("Cliente: " + gui.UserEntry.Text)
    gui.UserEntry.SetText("")
    if msgEnv == "bye" {
        channel <- true
        return
    }
}

func recMsgMain(socket net.Conn, channel chan bool, guiC *gui.GUIComponents, reports *gui.ReportComponents) { 
    reader := bufio.NewReader(socket)
    procIndex := 0
    for {
        socket.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
        msg, err := reader.ReadString('\n')
        if err != nil {
            if ne, ok := err.(net.Error); ok && ne.Timeout() {
                continue
            }
            fmt.Println("Conexión cerrada:", err)
            channel <- true
            return
        }
        if msg == "\r\n" || msg == "" {
            continue
        }
        fyne.Do(func() {
            var repPrefix, procPrefix bool
            msg, repPrefix = strings.CutPrefix(msg, "Report: ")
            msg, procPrefix = strings.CutPrefix(msg, "Process")
            switch {
            case repPrefix: //Estamos recibiendo un reporte
                cpu, ram, disk, ram_label, process_label, _ := mod_gui.ComponentValues(strings.TrimSpace(msg))
                reports.Cpu_Bar.SetValue(cpu/100.0)
                reports.Ram_Bar.SetValue(ram/100.0)
                reports.Disk_Bar.SetValue(disk/100.0)
                reports.Ram_Label.SetText(ram_label + " (Memoria usada / Memoria Total) ")
                reports.Process_Label.SetText("Cantidad de procesos en ejecución: " + process_label)
            case procPrefix: //Estamos recibiendo procesos
                cleanMsg := strings.TrimSpace(msg)
                if procIndex < len(reports.TopProcesses) {
                    reports.TopProcesses[procIndex].SetText(cleanMsg)
                    procIndex++
                }
                if procIndex >= len(reports.TopProcesses) {
                    procIndex = 0
                }
            default:
                guiC.AddMessage("Server: " + strings.TrimSpace(msg)) //Añade una label cada que se llame a la función
            }
        })
    }
}

//--------Llaves de comunicación entre los sockets-----------//
func recInvalidKey(socket net.Conn) string {
	msg, _ := bufio.NewReader(socket).ReadString('\n')
	fmt.Println("Server -->", msg)
	return msg
}

//--------Funciones bloqueantes para el flujo del login------//
func recDefaultMsg(socket net.Conn) {
	msg, _ := bufio.NewReader(socket).ReadString('\n')
	fmt.Println("Server -->", msg)
}

func sendCredentialMsg(socket net.Conn, credential string) {
	env := bufio.NewWriter(socket)
	env.WriteString(credential)
	env.Flush()
} 

//------Limpiar pantalla Windows-------//
func clearScreen() {
    shell := exec.Command("cmd", "/c", "cls")
    shell.Stdout = os.Stdout
    shell.Run()
}
