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
				phaseID := r.ID + "_exact"
				if !s.state.HasBeenSent(phaseID) {
					log.Printf("🚀 Enviando aviso EXACTO [ID: %s]: %s", phaseID, r.Content)

					// Si el webhook falla, no lo marcamos como enviado
					if err := s.webhook.Send(r, "🔴 [AHORA]"); err != nil {
						log.Printf("❌ Falló el envío del webhook para %s: %v", phaseID, err)
						continue
					}

					// Verificamos si realmente se guardó en el disco
					if err := s.state.MarkAsSent(phaseID); err != nil {
						log.Printf("🚨 ERROR FATAL: No se pudo guardar el estado de %s. Esto causará SPAM.", phaseID)
					}
					totalEnviados++
				}
				continue
			}

			// FASE 2: 1 HORA ANTES
			oneHourBefore := r.DueDate.Add(-1 * time.Hour)
			if now.After(oneHourBefore) {
				phaseID := r.ID + "_1h"
				if !s.state.HasBeenSent(phaseID) {
					log.Printf("🚀 Enviando aviso 1 HORA ANTES [ID: %s]: %s", phaseID, r.Content)

					if err := s.webhook.Send(r, "🟠 [EN 1 HORA]"); err == nil {
						s.state.MarkAsSent(phaseID)
						totalEnviados++
					}
				}
				continue
			}

			// FASE 1: HOY EN LA MAÑANA (9:00 AM)
			// Calculamos las 9:00 AM del mismo día del evento
			morning := time.Date(r.DueDate.Year(), r.DueDate.Month(), r.DueDate.Day(), 9, 0, 0, 0, time.Local)
			if r.DueDate.Hour() >= 10 && now.After(morning) {
				phaseID := r.ID + "_morning"
				if !s.state.HasBeenSent(phaseID) {
					log.Printf("🚀 Enviando aviso DE MAÑANA [ID: %s]: %s", phaseID, r.Content)

					if err := s.webhook.Send(r, "🟡 [PARA HOY]"); err == nil {
						s.state.MarkAsSent(phaseID)
						totalEnviados++
					}
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
