package main

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
)

var (
	filePath   = os.Args[1]
	fileExt    = filepath.Ext(filePath) // Récupérer l'extension de fichier
	withoutExt string
	ffmpegPath string
	// Définir les couleurs
	green   = color.New(color.FgGreen).SprintFunc()
	bold    = color.New(color.Bold)
	red     = color.New(color.FgRed)
	s       = spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	yellow  = color.New(color.FgYellow)
	crossed = color.New(color.FgRed).Add(color.CrossedOut).SprintFunc()
)

func main() {
	// Capture du signal d'interruption (Ctrl+C)
	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-interruptChannel
		yellow.Println("\rInterruption détectée. Arrêt en cours...")
		s.Stop()
		os.Exit(0)
	}()

	file, err := os.Stat(filePath)
	dir := filepath.Dir(filePath)
	startTime := time.Now()

	if err != nil {
		red.Println("\rUne erreur s'est produite\n", err, "\nVérifiez que le fichier existe.")
		yellow.Println("Si le chemin contient des espaces mettez le chemin entre guillemet.")
		os.Exit(0)
	}

	fileSize := float64(file.Size()) / (1024 * 1024) // convertir en MB
	fileSize = math.Round(fileSize*100) / 100        // arrondir

	// Détecter si la commande ffmpeg existe
	ffmpegPath, err = exec.LookPath("ffmpeg")
	if err != nil {
		red.Println("\rVous n'avez pas ffmpeg. Veuillez l'installer.")
		os.Exit(0)
	}
	withoutExt = file.Name()[0 : len(file.Name())-len(fileExt)]

	// Compresser une fois
	size := compressFile(filePath, 0)
	// En fonction du résultat il est possible que l'on doit réessayer
	for i := 1; i <= 3; i++ {
		if size > fileSize*(1-float64(i)*0.05) {
			// s.Suffix = "Réessaye avec d'autres paramètres...\nLe processus prendra plus de temps"
			size = compressFile(filePath, i)
		} else {
			break
		}
	}
	if size > fileSize {
		fmt.Println("Une erreur s'est produite : la taille du fichier compressé est supérieure à la taille initiale...")
		os.Exit(0)
	}
	bold.Println("Taille finale : ", crossed(fileSize, "MB"), "→", green(size, "MB"))
	originalFilePath := filepath.Join(dir, withoutExt+"_original"+fileExt)

	os.Rename(filePath, originalFilePath)
	os.Rename(filePath+".compressed"+fileExt, filePath)
	bold.Println("Votre vidéo se trouve ici : ", filePath)

	endTime := time.Now()
	elapsedTime := endTime.Sub(startTime)

	if elapsedTime < time.Minute {
		fmt.Printf("Temps d'exécution : %.2f secondes\n", elapsedTime.Seconds())
	} else {
		minutes := int(elapsedTime.Minutes())
		seconds := int(elapsedTime.Seconds()) - (minutes * 60)
		fmt.Printf("Temps d'exécution : %d minute.s %d seconde.s\n", minutes, seconds)
	}
}

func compressFile(filePath string, retryI int) float64 {
	var (
		parameter_crf = "28"
		parameter_r   = "60"
		parameter_b   = "500"
	)

	// En fonction du nombre d'essais, on change les paramètres
	if retryI == 1 {
		parameter_crf = "34"
		parameter_r = "50"
		parameter_b = "400"
	} else if retryI == 2 {
		parameter_crf = "38"
		parameter_r = "40"
		parameter_b = "300"
	} else if retryI == 3 {
		parameter_crf = "42"
		parameter_r = "30"
		parameter_b = "150"
	}

	s.Suffix = "  Compression en cours"
	s.Color("cyan")
	s.Start()
	cmd := exec.Command(
		ffmpegPath,
		"-hide_banner", "-loglevel", "error",
		"-i", filePath,
		"-vcodec", "libx264",
		"-preset", "slower",
		"-crf", parameter_crf,
		"-r", parameter_r,
		"-b", fmt.Sprintf("%sk", parameter_b),
		"-y", filePath+".compressed"+fileExt,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		red.Println("\r", err)
		os.Exit(0)
	}
	// Retourner la taille du fichier compressé
	file, err := os.Stat(filePath + ".compressed" + fileExt)
	if err != nil {
		red.Println("\rErreur lors de la récupération de la taille du fichier :", err)
		os.Exit(0)
	}

	fileSize := float64(file.Size()) / (1024 * 1024) // convertir en MB
	fileSize = math.Round(fileSize*100) / 100        // arrondir
	s.Stop()
	return fileSize
}
