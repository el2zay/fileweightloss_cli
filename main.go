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
	if len(os.Args) < 2 || os.Args[1] == "-logs" {
		fmt.Printf(`Usage: %s <chemin complet du fichier>

Flags:
     -logs : Active les logs de ffmpeg.
`, filepath.Base(os.Args[0]))
		// Quitter le programme proprement
		os.Exit(0)
	}

	filePath = os.Args[1]
	fileExt = filepath.Ext(filePath) // Récupérer l'extension de filePath

	flag.BoolVar(&ffmpegLogs, "logs", false, "Active les logs de ffmpeg.") // Définir le flag logs
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
			yellow.Println("Le fichier généré n'a pas pu être supprimé.")
		} else {
			yellow.Println("Le fichier", filePath+".compressed"+fileExt, "a bien été supprimer.")
		}
		os.Exit(0)
	}()

	file, err := os.Stat(filePath) // Récupérer des informations du fichier.
	dir := filepath.Dir(filePath)  // Récupérer le dossier du fichier
	startTime := time.Now()        // Commencer à compter en combien de temps le programme va s'executer

	// Si il y a une erreur lors de la vérification du fichier, avertir l'utilisateur`.
	if err != nil {
		red.Println("\rUne erreur s'est produite\n", err, "\nVérifiez que le fichier existe.")
		yellow.Println("Si le chemin contient des espaces, mettez-le entre guillemets.")
		s.Stop()
		os.Exit(0)
	}

	// filePath ne doit pas être un dossier.
	if file.IsDir() {
		red.Println("Veuillez spécifié un fichier.")
		os.Exit(0)
	}

	fileSize := float64(file.Size()) / (1024 * 1024) // récupérer la taille et convertir en MB
	fileSize = math.Round(fileSize*100) / 100        // arrondir

	// Détecter si la commande ffmpeg existe
	ffmpegPath, err = exec.LookPath("ffmpeg")
	// Sinon avertir l'utilisateur que ffmpeg n'est pas installé.
	if err != nil {
		red.Println("\rVous n'avez pas ffmpeg. Veuillez l'installer.")
		s.Stop()
		os.Exit(0)
	}
	withoutExt = file.Name()[0 : len(file.Name())-len(fileExt)] // Récupérer le nom du fichier mais sans son extension

	// Compresser une fois
	size := compressFile(filePath, 0)
	var i int
	// En fonction du résultat il est possible que l'on doit réessayer
	for i = 0; i <= 3; i++ {
		if size > fileSize*(1-float64(i)*0.05) {
			size = compressFile(filePath, i)
		} else {
			break
		}
	}

	// Si la taille du fichier compressé est plus grande que celle du fichier de base
	if size > fileSize {
		red.Println("\rUne erreur s'est produite : la taille du fichier compressé est supérieure à la taille initiale...")
		os.Exit(0)
	}
	// Si la taille du fichier compressé est égale à celle du fichier de base
	if size == fileSize {
		red.Println("\rLe fichier a peut-être déjà été compressé plusieurs fois, il ne peut donc pas perdre plus de données.")
	}

	// Afficher quelques informations.
	bold.Println("Taille finale : ", crossed(fileSize, "MB"), "→", green(size, "MB ", "(- ", math.Round(100-(100*size)/fileSize), " %)"))
	originalFilePath := filepath.Join(dir, withoutExt+"_original"+fileExt)

	os.Rename(filePath, originalFilePath) // Le fichier original prend un _original a son nom
	os.Rename(filePath+".compressed"+fileExt, filePath) // Le fichier compressé reprend le nom du fichier comme il était avant
	fmt.Println("Votre vidéo se trouve ici :", yellow.Sprint(filePath))

	endTime := time.Now() // Finir le temps ici
	elapsedTime := endTime.Sub(startTime) 

	// Si le programme a pris moins d'une minute pour se finir 
	if elapsedTime < time.Minute {
		fmt.Printf("Temps d'exécution : %.0f secondes\n", elapsedTime.Seconds()) // on l'affiche en secondes
	} else { //Sinon
		minutes := int(elapsedTime.Minutes())
		seconds := int(elapsedTime.Seconds()) - (minutes * 60)
		fmt.Printf("Temps d'exécution : %d minute.s et %d secondes\n", minutes, seconds) // on l'affiche en minute et secondes
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

	// S'il n'y a pas de flags -logs démarrer le spinner
	if !ffmpegLogs {
		s.Start()
	}

	// Executer la commande ffmpeg
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
