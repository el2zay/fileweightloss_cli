package main

import (
	"flag"
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
	filePath   string
	fileExt    string
	withoutExt string
	ffmpegPath string
	// Définir les couleurs
	green   = color.New(color.FgHiGreen).SprintFunc()
	bold    = color.New(color.Bold)
	red     = color.New(color.FgHiRed)
	s       = spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	yellow  = color.New(color.FgYellow)
	crossed = color.New(color.FgRed).Add(color.CrossedOut).SprintFunc()
	// Définir les flags
	ffmpegLogs bool
)

func main() {
	// Vérifier si filePath contient au moins un élément
	if len(os.Args) < 2 || os.Args[1] == "-logs"{
		fmt.Printf(`Usage: %s <chemin complet du fichier>

Flags:
     -logs : Active les logs de ffmpeg.
`, filepath.Base(os.Args[0]))
		// Quitter le programme proprement
		os.Exit(0)
	}

	filePath = os.Args[1]
	fileExt = filepath.Ext(filePath)

	flag.BoolVar(&ffmpegLogs, "logs", false, "Active les logs de ffmpeg.")
	flag.Parse()

	// Vérifier si le drapeau -logs est présent dans la ligne de commande
	for _, arg := range os.Args {
		if arg == "-logs" {
			ffmpegLogs = true
			break
		}
	}

	// Capture du signal d'interruption (Ctrl+C)
	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-interruptChannel
		yellow.Println("\rInterruption détectée. Arrêt en cours...")
		s.Stop() // Arreter le spinner pour éviter des problemes d'affichage dans le terminal.
		// Supprimer le fichier généré
		err := os.Remove(filePath + ".compressed" + fileExt)
		if err != nil {
			yellow.Println("Le fichier généré n'a a-pas pu être supprimé.")
		} else {
			yellow.Println("Le fichier", filePath+".compressed"+fileExt, "a bien été supprimer.")
		}
		os.Exit(0)
	}()

	file, err := os.Stat(filePath)
	dir := filepath.Dir(filePath)
	startTime := time.Now()

	if err != nil {
		red.Println("\rUne erreur s'est produite\n", err, "\nVérifiez que le fichier existe.")
		yellow.Println("Si le chemin contient des espaces, mettez-le entre guillemets.")
		s.Stop()
		os.Exit(0)
	}

	if file.IsDir() {
		red.Println("Veuillez spécifié un fichier.")
		os.Exit(0)
	}

	fileSize := float64(file.Size()) / (1024 * 1024) // convertir en MB
	fileSize = math.Round(fileSize*100) / 100        // arrondir

	// Détecter si la commande ffmpeg existe
	ffmpegPath, err = exec.LookPath("ffmpeg")
	if err != nil {
		red.Println("\rVous n'avez pas ffmpeg. Veuillez l'installer.")
		s.Stop()
		os.Exit(0)
	}
	withoutExt = file.Name()[0 : len(file.Name())-len(fileExt)]

	// Compresser une fois
	size := compressFile(filePath, 0)
	var i int
	// En fonction du résultat il est possible que l'on doit réessayer
	for i = 1; i <= 3; i++ {
		if size > fileSize*(1-float64(i)*0.05) {
			// s.Suffix = "Réessaye avec d'autres paramètres...\nLe processus prendra plus de temps"
			size = compressFile(filePath, i)
		} else {
			break
		}
	}

	if size > fileSize {
		red.Println("\rUne erreur s'est produite : la taille du fichier compressé est supérieure à la taille initiale...")
		os.Exit(0)
	}

	if size == fileSize {
		red.Println("\rLe fichier a peut-être déjà été compressé plusieurs fois, il ne peut donc pas perdre plus de données.")
	}

	bold.Println("Taille finale : ", crossed(fileSize, "MB"), "→", green(size, "MB ", "(- ", math.Round(100-(100*size)/fileSize), " %)"))
	originalFilePath := filepath.Join(dir, withoutExt+"_original"+fileExt)

	os.Rename(filePath, originalFilePath)
	os.Rename(filePath+".compressed"+fileExt, filePath)
	fmt.Println("Votre vidéo se trouve ici :", yellow.Sprint(filePath))

	endTime := time.Now()
	elapsedTime := endTime.Sub(startTime)

	if elapsedTime < time.Minute {
		fmt.Printf("Temps d'exécution : %.0f secondes\n", elapsedTime.Seconds())
	} else {
		minutes := int(elapsedTime.Minutes())
		seconds := int(elapsedTime.Seconds()) - (minutes * 60)
		fmt.Printf("Temps d'exécution : %d minute.s et %d secondes\n", minutes, seconds)
	}
	fmt.Println("Nombre de tentative : ", bold.Sprint(i+1))

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
	// Déclarer une variable pour stocker la valeur du drapeau -logs

	// Analyser les drapeaux de la ligne de commande
	flag.Parse()

	if !ffmpegLogs {
		s.Start()
	}

	cmdArgs := []string{
		ffmpegPath,
		"-i", filePath,
		"-vcodec", "libx264",
		"-preset", "slower",
		"-crf", parameter_crf,
		"-r", parameter_r,
		"-b", fmt.Sprintf("%sk", parameter_b),
		"-y", filePath + ".compressed" + fileExt,
	}

	if !ffmpegLogs {
		cmdArgs = append(cmdArgs, "-hide_banner", "-loglevel", "error")
	}

	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println("Erreur lors de l'exécution de la commande:", err)
	}

	if err != nil {
		red.Println(err)
		s.Stop()
		os.Exit(0)
	}
	// Retourner la taille du fichier compressé
	file, err := os.Stat(filePath + ".compressed" + fileExt)
	if err != nil {
		red.Println("\rErreur lors de la récupération de la taille du fichier :", err)
		s.Stop()
		os.Exit(0)
	}

	fileSize := float64(file.Size()) / (1024 * 1024) // convertir en MB
	fileSize = math.Round(fileSize*100) / 100        // arrondir
	s.Stop()
	return fileSize
}
