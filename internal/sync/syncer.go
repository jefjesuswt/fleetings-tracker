package sync

import (
	"log"
	"time"

	"github.com/jefjesuswt/fleetings-tracker/internal/github"
	"github.com/jefjesuswt/fleetings-tracker/internal/parser"
	"github.com/jefjesuswt/fleetings-tracker/internal/state"
	"github.com/jefjesuswt/fleetings-tracker/internal/webhook"
)

type Syncer struct {
	gitClient    *github.Client
	webhook      *webhook.Client
	state        *state.Store
	obsidianPath string
}

func NewSyncer(git *github.Client, wh *webhook.Client, st *state.Store, path string) *Syncer {
	return &Syncer{
		gitClient:    git,
		webhook:      wh,
		state:        st,
		obsidianPath: path,
	}
}

func (s *Syncer) Run() error {
	files, err := s.gitClient.ListFleetings(s.obsidianPath)
	if err != nil {
		return err
	}

	totalEnviados := 0
	now := time.Now() // Tomamos la hora actual exacta

	for _, file := range files {
		content, err := s.gitClient.GetFileContent(file)
		if err != nil {
			continue
		}

		reminders := parser.ExtractReminders(file, content)

		for _, r := range reminders {

			// FASE 3: HORA EXACTA (Ya pasó o es el momento justo)
			if now.After(r.DueDate) {
				if !s.state.HasBeenSent(r.ID + "_exact") {
					log.Printf("🚀 Enviando aviso EXACTO: %s", r.Content)
					s.webhook.Send(r, "🔴 [AHORA]")
					s.state.MarkAsSent(r.ID + "_exact")
					totalEnviados++
				}
				continue // Si ya pasó la hora exacta, no evaluamos las fases anteriores
			}

			// FASE 2: 1 HORA ANTES
			oneHourBefore := r.DueDate.Add(-1 * time.Hour)
			if now.After(oneHourBefore) {
				if !s.state.HasBeenSent(r.ID + "_1h") {
					log.Printf("🚀 Enviando aviso 1 HORA ANTES: %s", r.Content)
					s.webhook.Send(r, "🟠 [EN 1 HORA]")
					s.state.MarkAsSent(r.ID + "_1h")
					totalEnviados++
				}
				continue
			}

			// FASE 1: HOY EN LA MAÑANA (9:00 AM)
			// Calculamos las 9:00 AM del mismo día del evento
			morning := time.Date(r.DueDate.Year(), r.DueDate.Month(), r.DueDate.Day(), 9, 0, 0, 0, time.Local)

			// Para que aplique, el evento debe ser después de las 10:00 AM
			// (Para que no se envíe al mismo tiempo que el aviso de "1 hora antes")
			if r.DueDate.Hour() >= 10 && now.After(morning) {
				if !s.state.HasBeenSent(r.ID + "_morning") {
					log.Printf("🚀 Enviando aviso DE MAÑANA: %s", r.Content)
					s.webhook.Send(r, "🟡 [PARA HOY]")
					s.state.MarkAsSent(r.ID + "_morning")
					totalEnviados++
				}
			}
		}
	}

	if totalEnviados > 0 {
		log.Printf("✅ Se enviaron %d notificaciones.", totalEnviados)
	} else {
		log.Println("💤 No hay reminders nuevos pendientes.")
	}

	return nil
}
