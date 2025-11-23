// Aqui unicamente maneja el layout de la gui del programa, no edita realmente
package gui

import (
	"time"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type GUIComponents struct {
    SendButton *widget.Button //Necesario para que al accionar se pueda enviar el mensaje al socket servidor
    MsgBox *fyne.Container //Contenedor de la caja de mensajes, necesario para agregar labels de los mensajes que recibe el cliente por parte del servidor
    UserEntry *widget.Entry //Input del usuario en el sentido de mensajes
    ScrollOfBox *container.Scroll
    autoScroll bool
}

type ReportComponents struct {
    //Labels
    Ram_Label *widget.Label
    Process_Label *widget.Label
    TopProcesses []*widget.Label
    //Bar
    Cpu_Bar *widget.ProgressBar
    Ram_Bar *widget.ProgressBar
    Disk_Bar *widget.ProgressBar
}

func StartMainWindow(app fyne.App) (*fyne.Window, *GUIComponents, *ReportComponents) { //Layout de la ventana principal
    mainWin := app.NewWindow("Operclite App")
    topBar := widget.NewLabel("Operclite Beta - Escribe 'exec: <tu comando>' si quieres ejecutar un comando - Escribe 'bye' en el panel de comunicación si quieres terminar la conexión")
    topBar.Alignment = fyne.TextAlignCenter
    
    //==== Caja de labels para el top de procesos ======
    topProcesses := []*widget.Label {
        widget.NewLabel(""),
        widget.NewLabel(""),
        widget.NewLabel(""),
        widget.NewLabel(""),
        widget.NewLabel(""),
        widget.NewLabel(""),
    }

    //======= CPU ===========
    cpu_label := widget.NewLabel("Porcentaje de uso del CPU")
    cpu_bar := widget.NewProgressBar()
    //======= RAM ===========
    ram_label := widget.NewLabel("")
    ram_bar := widget.NewProgressBar()
    //======= Disk ==========
    disk_label := widget.NewLabel("Porcentaje de espacio ocupado en disco")
    disk_bar := widget.NewProgressBar()

    //Procesos
    process_label := widget.NewLabel("Cantidad de procesos en ejecución")

    //Alineaciones de texto
    ram_label.Alignment = fyne.TextAlignCenter
    cpu_label.Alignment = fyne.TextAlignCenter
    disk_label.Alignment = fyne.TextAlignCenter
    process_label.Alignment = fyne.TextAlignCenter
    
    // ========= Contenedor de reportes ============
    realTimeReport := container.NewVBox(
        cpu_label,
        cpu_bar,
        ram_label,
        ram_bar,
        disk_label,
        disk_bar,
        process_label,
    )

    // ======= Labels decoradores ===============
    topProcessLabel := widget.NewLabel(" ==========// Procesos de mayor consumo del servidor //========== ")
    topProcessLabel.Alignment = fyne.TextAlignCenter

    reportLabel := widget.NewLabel(" ===========// Detalle de consumo de recursos del servidor //========= ")
    reportLabel.Alignment = fyne.TextAlignCenter
    
    // ======= Centrar labels de procesos ========
    for _, label := range topProcesses {
        label.Alignment = fyne.TextAlignCenter 
    }
     
    //Caja completa de la izquierda
    fullReportContainer := container.NewVBox( 
        topProcessLabel,
        topProcesses[0],
        topProcesses[1],
        topProcesses[2],
        topProcesses[3],
        topProcesses[4],
        topProcesses[5],
        reportLabel,
        realTimeReport,
    )

    //Layout de la caja de mensajes con scroll
    msgBox := container.NewVBox()
    scroll_of_box := container.NewVScroll(msgBox)
    scroll_of_box.SetMinSize(fyne.NewSize(612, 612))
    
    
    //Caja de entrada de mensajes
    userEntry := widget.NewEntry()
    userEntry.SetPlaceHolder("Escribe tu mensaje aquí...") 
    sendButton := widget.NewButton("Enviar", func() {})
    bottomBar := container.NewAdaptiveGrid(2, userEntry, sendButton) 

    // ====== Caja de comunicacion, aquí los sockets intercambian mensajes
    communicationInterfacePanel := container.NewBorder(
        scroll_of_box, //Caja Top
        bottomBar, //Zona de escribir y enviar mensaje [Caja Bottom] ------container.NewBorder(nil, nil, userEntry, sendButton,nil)
        nil, //Caja Left
        nil, //Caja Right
        nil, //Center
    )
    centerPanel := container.NewAdaptiveGrid(2, fullReportContainer, communicationInterfacePanel)

    //Layout completo
    fullLayout := container.NewBorder(topBar, nil, nil, nil, centerPanel) 
    mainWin.SetContent(fullLayout)
    
    //Estructura de los componentes generales de la GUI
    gui := &GUIComponents{
        SendButton: sendButton,
        MsgBox: msgBox,
        UserEntry: userEntry,
        ScrollOfBox: scroll_of_box,
        autoScroll: true,
    }

    scroll_of_box.OnScrolled = func(pos fyne.Position) {
        contentHeight := msgBox.MinSize().Height
        visibleHeight := scroll_of_box.Size().Height
        scrollPos := pos.Y
        gui.autoScroll = (scrollPos + visibleHeight) >= (contentHeight - 50)
    }

    //Estructura de los componentes internos de la caja de reportes (5 procesos de mayor consumo, consumo general del sistema)
    report := &ReportComponents{
        Ram_Label: ram_label,
        Process_Label: process_label,
        TopProcesses: topProcesses,
        Cpu_Bar: cpu_bar,
        Ram_Bar: ram_bar,
        Disk_Bar: disk_bar,
    }

    return &mainWin, gui, report
}

//Función para añadir mensajes a la caja de mensajes aplicando un scroll automático
func (gui *GUIComponents) AddMessage(msg string) {
    label := widget.NewLabel(msg)
    gui.MsgBox.Add(label)
    gui.MsgBox.Refresh()
    if gui.autoScroll {
        time.AfterFunc(10 * time.Millisecond, func() {
            fyne.Do(func() {gui.ScrollOfBox.ScrollToBottom()})
        })
    }
}

 //Ejecuta la ventana de error
func ErrorWindow(app fyne.App) {
    errorWin := app.NewWindow("Operclite error")
    errorLabel := widget.NewLabel("Error, acceso no autorizado, imposible conectarse.")
    closeBTN := widget.NewButton("Cerrar", func() {
        errorWin.Close()
    })
    errorWin.SetContent(container.NewVBox(
        errorLabel,
        closeBTN,
    ))
    errorWin.ShowAndRun()
}

//Ejecuta la ventana de despedida
func ByeWindow(app fyne.App) { 
    byeWin := app.NewWindow("Operclite Closed")
    byeLabel := widget.NewLabel("Gracias por usar el cliente Operclite :)")
    closeBTN := widget.NewButton("Cerrar", func() {
        byeWin.Close()
    })
    byeWin.SetContent(container.NewVBox(
        byeLabel,
        closeBTN,
    ))
    byeWin.Show()
}