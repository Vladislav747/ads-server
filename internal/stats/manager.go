package stats

import (
	"context"
	"log"
	"sync"
	"time"
)

const (
	// Размер начальной ёмкости map для статистики.
	defaultCapacity = 1000
)

type (
	// Writer — интерфейс хранилища статистики.
	Writer interface {
		Insert(rows Rows) error
	}
	// Manager агрегирует статистику и периодически пишет её через Writer.
	Manager struct {
		writer        Writer
		flushInterval time.Duration

		ctx    context.Context
		cancel context.CancelFunc

		// Мьютекс защищает доступ к rows.
		mu   sync.Mutex
		rows Rows
	}
)

// NewManager создаёт менеджер с фоновым контекстом и пустым набором строк.
func NewManager(w Writer, flushInterval time.Duration) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	return &Manager{
		writer:        w,
		flushInterval: flushInterval,
		ctx:           ctx,
		cancel:        cancel,
		rows:          newRows(),
	}
}

// Append добавляет (или агрегирует) одну пару ключ–значение.
func (m *Manager) Append(k Key, v Value) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.unsafeAppend(k, v)
}

// AppendRows добавляет множество строк сразу.
func (m *Manager) AppendRows(rows Rows) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for k, v := range rows {
		m.unsafeAppend(k, v)
	}
}

func (m *Manager) unsafeAppend(k Key, v Value) {
	current := m.rows[k]
	current = current.Assign(v)

	m.rows[k] = current
}

// Start запускает фоновый цикл записи статистики.
func (m *Manager) Start() {
	log.Println("Stats loop started")
	go m.loop()
}

// loop периодически инициирует запись статистики или завершает работу по ctx.
func (m *Manager) loop() {
	for {
		select {
		case <-time.After(m.flushInterval):
			m.startInserting()

		case <-m.ctx.Done():
			m.startInserting()
			return
		}
	}
}

// startInserting забирает накопленные строки и пытается записать их через Writer.
func (m *Manager) startInserting() {
	log.Println("Start stats inserting")

	rows := m.withdraw()
	if len(rows) == 0 {
		log.Println("No stats rows, skipping")
		return
	}

	if err := m.writer.Insert(rows); err != nil {
		log.Printf("Failed to write stats: %v\n", err)
		log.Printf("Return stats rows to map: %d\n", len(rows))

		m.AppendRows(rows)
		return
	}

	log.Printf("Stats rows successfuly written: %d\n", len(rows))
}

// withdraw атомарно забирает текущие rows и заменяет их на новый map.
func (m *Manager) withdraw() Rows {
	m.mu.Lock()
	defer m.mu.Unlock()

	rows := m.rows
	m.rows = newRows()

	return rows
}

// newRows создаёт map с предвыделенной ёмкостью.
func newRows() Rows {
	return make(Rows, defaultCapacity)
}
