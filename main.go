package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jefjesuswt/fleetings-tracker/internal/github"
	"github.com/jefjesuswt/fleetings-tracker/internal/parser"
	"github.com/joho/godotenv"
)

func main() {
	log.Println("⚙️ [INIT] Cargando variables de entorno...")
	err := godotenv.Load()
	if err != nil {
		log.Println("ℹ️ [INFO] No se encontró archivo .env, leyendo variables de entorno del sistema...")
	}

	gitToken := os.Getenv("GIT_TOKEN")
	gitOwner := os.Getenv("GIT_OWNER")
	gitRepo := os.Getenv("GIT_REPO")
	obsidianFleetingPath := os.Getenv("OBSIDIAN_FLEETING_PATH")
	webhookUrl := os.Getenv("WEBHOOK_URL")

	if obsidianFleetingPath == "" {
		obsidianFleetingPath = "05 - Fleetings/Sticky Reminders"
		log.Printf("⚠️ [WARN] OBSIDIAN_FLEETING_PATH vacío. Usando por defecto: '%s'", obsidianFleetingPath)
	}

	if gitToken == "" || gitOwner == "" || gitRepo == "" {
		log.Fatal("❌ [FATAL] Faltan variables de entorno")
	}

	requiredVars := map[string]string{
		"GIT_TOKEN":   gitToken,
		"GIT_OWNER":   gitOwner,
		"GIT_REPO":    gitRepo,
		"WEBHOOK_URL": webhookUrl,
	}

	for key, value := range requiredVars {
		if value == "" {
			log.Fatalf("❌ [FATAL] Falta variable de entorno: %s", key)
		}
	}

	log.Println("✅ [INIT] Variables cargadas correctamente. Inicializando componentes...")

	gitClient := github.NewClient(gitToken, gitOwner, gitRepo)

	log.Println("✅ [INIT] Componentes inicializados correctamente. Iniciando servidor...")

	files, err := gitClient.ListFleetings(obsidianFleetingPath)
	if err != nil {
		log.Fatal(err)
	}

	if len(files) == 0 {
		log.Println("ℹ️ [INFO] No hay fleetings en el archivo de fleetings especificado")
		return
	}

	for index, file := range files {
		fmt.Printf(" %d. %s\n", index+1, file)
	}

	firstFile := files[0]
	fmt.Printf("ℹ️ [INFO] El primer fleeting es: %s\n", firstFile)

	content, err := gitClient.GetFileContent(firstFile)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("✅ [TEST] content descargado con éxito. Primeros 200 caracteres:")
	fmt.Println("--------------------------------------------------")

	if len(content) > 200 {
		fmt.Println(content[:200] + "...\n[CONTINÚA]")
	} else {
		fmt.Println(content)
	}
	fmt.Println("--------------------------------------------------")

	log.Println("🧠 [TEST] Ejecutando parser de expresiones regulares...")

	contentWithReminders := content + "\n- [ ] Actualizar paquetes de Arch Linux [webhook@2026-03-05 18:30]"

	reminders := parser.ExtractReminders(firstFile, contentWithReminders)

	if len(reminders) == 0 {
		log.Println("⚠️ [TEST] No se encontraron reminders.")
	} else {
		log.Printf("✅ [TEST] Se encontraron %d reminders:", len(reminders))
		for _, r := range reminders {
			fmt.Printf("   -> ID: %s\n", r.ID)
			fmt.Printf("   -> Tarea: %s\n", r.Content)
			fmt.Printf("   -> Fecha: %s\n", r.DueDate.Format("02/01/2006 a las 15:04"))
			fmt.Printf("   -> Archivo: %s\n", r.File)
		}
	}
}
