package state

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"
)

type Store struct {
	filePath string
	mu       sync.Mutex
	sentIDs  map[string]time.Time
}

func NewStore(filePath string) (*Store, error) {
	s := &Store{
		filePath: filePath,
		sentIDs:  make(map[string]time.Time),
	}
	s.load()
	return s, nil
}

func (s *Store) load() {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		log.Printf("ℹ️ [STATE] No se pudo leer '%s' (Es normal si es la primera vez): %v", s.filePath, err)
		return
	}

	if err := json.Unmarshal(data, &s.sentIDs); err != nil {
		log.Printf("⚠️ [STATE] El archivo '%s' está vacío o corrupto. Se reiniciará. Error: %v", s.filePath, err)
	}

	// Protección vital: Si el archivo JSON tenía un "null", Go convierte el mapa en nil.
	if s.sentIDs == nil {
		s.sentIDs = make(map[string]time.Time)
	}

	log.Printf("✅ [STATE] Memoria cargada: %d recordatorios previos detectados.", len(s.sentIDs))
}

func (s *Store) save() error {
	data, err := json.MarshalIndent(s.sentIDs, "", "  ")
	if err != nil {
		log.Printf("❌ [STATE] Error serializando el JSON: %v", err)
		return err
	}

	err = os.WriteFile(s.filePath, data, 0644)
	if err != nil {
		log.Printf("🚨 [STATE] ERROR CRÍTICO: No se tienen permisos para escribir en '%s': %v", s.filePath, err)
	} else {
		log.Printf("💾 [STATE] Archivo guardado correctamente en disco.")
	}

	return err
}

func (s *Store) HasBeenSent(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, exists := s.sentIDs[id]
	return exists
}

func (s *Store) MarkAsSent(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.sentIDs == nil {
		s.sentIDs = make(map[string]time.Time)
	}

	s.sentIDs[id] = time.Now()
	return s.save() // Disparamos el guardado inmediatamente
}
