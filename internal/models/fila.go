package models

import "time"

// Enums para Fila
type PrioridadeFila string
const (
	PrioridadeFilaBaixa   PrioridadeFila = "BAIXA"
	PrioridadeFilaMedia   PrioridadeFila = "MEDIA"
	PrioridadeFilaAlta    PrioridadeFila = "ALTA"
	PrioridadeFilaUrgente PrioridadeFila = "URGENTE"
)

// Fila representa uma fila de atendimento
type Fila struct {
	BaseModel
	Nome          string         `gorm:"not null" json:"nome"`
	Descricao     string         `gorm:"not null" json:"descricao"`
	Cor           string         `gorm:"not null" json:"cor"`
	Ordenacao     int            `gorm:"not null" json:"ordenacao"`
	Ativa         bool           `gorm:"default:true" json:"ativa"`
	Prioridade    PrioridadeFila `gorm:"default:MEDIA" json:"prioridade"`
	ChatBot       bool           `gorm:"default:false" json:"chatBot"`
	Kanban        bool           `gorm:"default:false" json:"kanban"`
	WhatsappChats bool           `gorm:"default:true" json:"whatsappChats"`

	// Relacionamentos
	Atendentes []FilaAtendente `gorm:"foreignKey:FilaID" json:"atendentes,omitempty"`
}

// FilaAtendente representa a relação many-to-many entre Fila e Usuario
type FilaAtendente struct {
	ID        string    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	FilaID    string    `gorm:"type:uuid;not null" json:"filaId"`
	UsuarioID string    `gorm:"type:uuid;not null" json:"usuarioId"`
	CriadoEm  time.Time `gorm:"autoCreateTime" json:"criadoEm"`

	// Relacionamentos
	Fila    Fila    `gorm:"foreignKey:FilaID" json:"fila,omitempty"`
	Usuario Usuario `gorm:"foreignKey:UsuarioID" json:"usuario,omitempty"`
}

// FilaRequest para criação/atualização de fila
type FilaRequest struct {
	Nome          string         `json:"nome" validate:"required,min=3,max=100"`
	Descricao     string         `json:"descricao" validate:"required,min=10,max=500"`
	Cor           string         `json:"cor" validate:"required,hexcolor"`
	Ordenacao     int            `json:"ordenacao" validate:"required,min=1"`
	Ativa         bool           `json:"ativa"`
	Prioridade    PrioridadeFila `json:"prioridade" validate:"required,oneof=BAIXA MEDIA ALTA URGENTE"`
	ChatBot       bool           `json:"chatBot"`
	Kanban        bool           `json:"kanban"`
	WhatsappChats bool           `json:"whatsappChats"`
	AtendentesIDs []string       `json:"atendentesIds"`
}

// FilaWithStats representa uma fila com estatísticas
type FilaWithStats struct {
	Fila
	Estatisticas FilaEstatisticas `json:"estatisticas"`
}

// FilaEstatisticas representa as estatísticas de uma fila
type FilaEstatisticas struct {
	TotalConversas      int     `json:"totalConversas"`
	ConversasAtivas     int     `json:"conversasAtivas"`
	TempoMedioResposta  float64 `json:"tempoMedioResposta"` // em minutos
	Satisfacao          float64 `json:"satisfacao"`        // de 0 a 5
}

// TableName especifica o nome da tabela
func (Fila) TableName() string {
	return "filas"
}

func (FilaAtendente) TableName() string {
	return "fila_atendentes"
}
