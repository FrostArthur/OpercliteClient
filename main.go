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
    "fyne.io/fyne/v2/widget"
)

func main() {
    //Logica de interconexion
    var operclite fyne.App = app.New() //Variable de la aplicacion
    clearScreen()
    //ptoRemoto := "1610"
    //ipRemota := "10.1.6.197"
    fmt.Print("Ingrese el puerto remoto por el que va a ingresar: ")
    portReader := bufio.NewReader(os.Stdin)
	ptoRemoto, err := portReader.ReadString('\n')
    if err != nil {
        fmt.Println("Error de lectura", err)
        os.Exit(0)
    }
    fmt.Print("Ingrese la IP del socket servidor (OJO, RECUERDE ESTAR BAJO LA MISMA RED): ")
    ipReader := bufio.NewReader(os.Stdin)
	ipRemota, err := ipReader.ReadString('\n')
    if err != nil {
        fmt.Println("Error de lectura", err)
        os.Exit(0)
    }
    dirTCP, _ := net.ResolveTCPAddr("tcp4", strings.TrimSpace(ipRemota)+":"+strings.TrimSpace(ptoRemoto))
    var serverSocket net.Conn
    serverSocket, _ = net.DialTCP("tcp4", nil, dirTCP)
	//Recibir mensaje por defecto de el servidor
	if recInvalidKey(serverSocket) == "Unauthorized\r\n" { //Conexion invalida
        gui.ErrorWindow(operclite)
		serverSocket.Close()
		os.Exit(0)
	}
    fmt.Println("\n--------------------------------------")
    fmt.Println("*           Operclite Client           *")
    fmt.Println("----------------------------------------")
    //Proceso de login == CLI para poder implementar mejor seguridad a la hora de escribir la contraseña
	for {
		//var usercli string
	    recDefaultMsg(serverSocket) //Recibe mensaje de ingresar usuario
		fmt.Print("Ingrese el usuario: ")
		reader := bufio.NewReader(os.Stdin)
        usercli, _ := reader.ReadString('\n')
        sendCredentialMsg(serverSocket, usercli) //Envia la credencial

		recDefaultMsg(serverSocket) //Recibe mensaje de ingresar contraseña
		fmt.Print("Ingrese la contraseña: ")
		hPasswd, err := readPassword()
	    if err != nil {
		    fmt.Println("Error:", err)
		    os.Exit(1)
	    }
		sendCredentialMsg(serverSocket, hPasswd) //Envia la credencial
		tempKey := recInvalidKey(serverSocket)
		if tempKey == "Login OK\r\n" {
            fmt.Println("Login correcto, iniciando comunicacion...")
		    break
		} else if tempKey == "Locked\r\n" {
            clearScreen()
			serverSocket.Close()
			gui.ErrorWindow(operclite)
			os.Exit(0) 
	    }
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
            (*mainWin).Close()
            fmt.Println("Bye :)")
            //os.Exit(0)
        })
    }()
    (*mainWin).SetIcon(resourceIconPng)
    (*mainWin).ShowAndRun() //Inicia la ventana principal
    gui.ByeWindow(operclite)
    //clearScreen()
    /*fmt.Println("\n--------------------------------------------")
    fmt.Println("*  Gracias por usar el cliente de Operclite  *")
    fmt.Println("----------------------------------------------")*/
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
    gui.MsgBox.Add(widget.NewLabel("Cliente: " + gui.UserEntry.Text))
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
        if strings.TrimSpace(msg) == "bye" {
            channel <- true
            return
        }
        if msg == "\n" || msg == "" {
            continue
        }
        fyne.Do(func() {
            if strings.HasPrefix(msg, "Report: ") {
                cpu, ram, disk, ram_label, process_label, err := mod_gui.ComponentValues(strings.TrimPrefix(msg, "Report: "))
                if err != nil {
                    fmt.Println(err)
                    return
                }
                reports.Cpu_Bar.SetValue(cpu/100.0)
                reports.Ram_Bar.SetValue(ram/100.0)
                reports.Disk_Bar.SetValue(disk/100.0)
                reports.Ram_Label.SetText(ram_label + " (Memoria usada / Memoria Total) ")
                reports.Process_Label.SetText("Cantidad de procesos en ejecución: " + process_label) 
            } else if strings.HasPrefix(msg, "Process") {
                cleanMsg := strings.TrimSpace(strings.TrimPrefix(msg, "Process"))
                if procIndex < len(reports.TopProcesses) {
                    reports.TopProcesses[procIndex].SetText(cleanMsg)
                    procIndex++
                }
                if procIndex >= len(reports.TopProcesses) {
                    procIndex = 0
                }
            } else {
                guiC.MsgBox.Add(widget.NewLabel("Server: " + strings.TrimSpace(msg))) //Añade una label cada que se llame a la función
            }   
        })
    }
}

//--------Llave de comunicación entre los sockets-----------//

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
