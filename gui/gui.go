// Aqui unicamente maneja el layout de la gui del programa, no edita realmente
package gui

import (
	"time"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type GUIComponents struct {
    SendButton *widget.Button
    MsgBox *fyne.Container 
    UserEntry *widget.Entry
    ScrollOfBox *container.Scroll
    autoScroll bool
}

type ReportComponents struct {
    Ram_Label *widget.Label
    Process_Label *widget.Label
    TopProcesses []*widget.Label
    Cpu_Bar *widget.ProgressBar
    Ram_Bar *widget.ProgressBar
    Disk_Bar *widget.ProgressBar
}

func StartMainWindow(app fyne.App) (*fyne.Window, *GUIComponents, *ReportComponents) { 
    mainWin := app.NewWindow("Operclite App")
    topBar := widget.NewLabel("Operclite Beta - Escribe 'exec: <tu comando>' si quieres ejecutar un comando - Escribe 'bye' en el panel de comunicación si quieres terminar la conexión")
    topBar.Alignment = fyne.TextAlignCenter
    
    topProcesses := []*widget.Label {
        widget.NewLabel(""),
        widget.NewLabel(""),
        widget.NewLabel(""),
        widget.NewLabel(""),
        widget.NewLabel(""),
        widget.NewLabel(""),
    }

    cpu_label := widget.NewLabel("Porcentaje de uso del CPU")
    cpu_bar := widget.NewProgressBar()
    ram_label := widget.NewLabel("")
    ram_bar := widget.NewProgressBar()
    disk_label := widget.NewLabel("Porcentaje de espacio ocupado en disco")
    disk_bar := widget.NewProgressBar()

    process_label := widget.NewLabel("Cantidad de procesos en ejecución")

    ram_label.Alignment = fyne.TextAlignCenter
    cpu_label.Alignment = fyne.TextAlignCenter
    disk_label.Alignment = fyne.TextAlignCenter
    process_label.Alignment = fyne.TextAlignCenter
    
    realTimeReport := container.NewVBox(
        cpu_label,
        cpu_bar,
        ram_label,
        ram_bar,
        disk_label,
        disk_bar,
        process_label,
    )

    topProcessLabel := widget.NewLabel(" ==========// Procesos de mayor consumo del servidor //========== ")
    topProcessLabel.Alignment = fyne.TextAlignCenter

    reportLabel := widget.NewLabel(" ===========// Detalle de consumo de recursos del servidor //========= ")
    reportLabel.Alignment = fyne.TextAlignCenter
    
    for _, label := range topProcesses {
        label.Alignment = fyne.TextAlignCenter 
    }
     
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

    msgBox := container.NewVBox()
    scroll_of_box := container.NewVScroll(msgBox)
    scroll_of_box.SetMinSize(fyne.NewSize(612, 612))
    
    userEntry := widget.NewEntry()
    userEntry.SetPlaceHolder("Escribe tu mensaje aquí...") 
    sendButton := widget.NewButton("Enviar", func() {})
    bottomBar := container.NewAdaptiveGrid(2, userEntry, sendButton) 

    communicationInterfacePanel := container.NewBorder(
        scroll_of_box,
        bottomBar,
        nil,
        nil,
        nil,
    )
    centerPanel := container.NewAdaptiveGrid(2, fullReportContainer, communicationInterfacePanel)

    fullLayout := container.NewBorder(topBar, nil, nil, nil, centerPanel) 
    mainWin.SetContent(fullLayout)
    
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