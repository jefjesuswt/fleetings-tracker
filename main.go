package main

import (
	"log"
	"os"

	"github.com/jefjesuswt/fleetings-tracker/internal/github"
	"github.com/jefjesuswt/fleetings-tracker/internal/state"
	"github.com/jefjesuswt/fleetings-tracker/internal/sync"
	"github.com/jefjesuswt/fleetings-tracker/internal/webhook"
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
	whClient := webhook.NewClient(webhookUrl)
	stateStore, err := state.NewStore("sent_reminders.json")
	if err != nil {
		log.Fatalf("❌ [FATAL] Error inicializando el estado local: %v", err)
	}

	log.Println("✅ [INIT] Componentes inicializados correctamente. Iniciando el sincronizador...")

	syncer := sync.NewSyncer(gitClient, whClient, stateStore, obsidianFleetingPath)

	log.Println("✅ [INIT] Todo listo para ejecutar el sincronizador. Iniciando...")
	if err := syncer.Run(); err != nil {
		log.Fatalf("❌ [FATAL] Error ejecutando el sincronizador: %v", err)
	}

	log.Println("--------------------------------------------------")
	log.Println("🎉 [DONE] Escaneo completado. Esperando el próximo ciclo del cron.")
}
