package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// JSONB type para campos JSON no PostgreSQL
type JSONB map[string]interface{}

func (j JSONB) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONB)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, j)
}

type Tag struct {
	BaseModel
	Nome       string `gorm:"uniqueIndex;not null" json:"nome"`
	Descricao  string `gorm:"type:text" json:"descricao,omitempty"`
	Cor        string `gorm:"not null;default:'#3b82f6'" json:"cor"`
	Categoria  string `gorm:"not null;default:'geral'" json:"categoria"`
	UsoCount   int    `gorm:"default:0" json:"uso_count"`
	CriadoPor  string `gorm:"not null" json:"criado_por"`
	Ativo      bool   `gorm:"default:true" json:"ativo"`
	Favorito   bool   `gorm:"default:false" json:"favorito"`

	// Relacionamentos
	Contatos []ContatoTag `gorm:"foreignKey:TagID" json:"contatos,omitempty"`
	Quadros  []QuadroTag  `gorm:"foreignKey:TagID" json:"quadros,omitempty"`
}

func (Tag) TableName() string {
	return "tags"
}

type ContatoTag struct {
	BaseModel
	ContatoID string `gorm:"not null" json:"contatoId"`
	TagID     string `gorm:"not null" json:"tagId"`

	// Relacionamentos
	Contato Contato `gorm:"foreignKey:ContatoID" json:"contato,omitempty"`
	Tag     Tag     `gorm:"foreignKey:TagID" json:"tag,omitempty"`
}

func (ContatoTag) TableName() string {
	return "contato_tags"
}

type Quadro struct {
	BaseModel
	Nome      string  `gorm:"not null" json:"nome"`
	Cor       string  `gorm:"not null" json:"cor"`
	Descricao *string `json:"descricao"`
	Posicao   int     `gorm:"not null" json:"posicao"`
	Ativo     bool    `gorm:"default:true" json:"ativo"`
	UsuarioID string  `gorm:"column:usuario_id;not null" json:"usuarioId"`

	// Relacionamentos
	Usuario Usuario     `gorm:"foreignKey:UsuarioID" json:"usuario,omitempty"`
	Colunas []Coluna    `gorm:"foreignKey:QuadroID" json:"colunas,omitempty"`
	Tags    []QuadroTag `gorm:"foreignKey:QuadroID" json:"tags,omitempty"`
	Fluxos  []Fluxo     `gorm:"foreignKey:QuadroID" json:"fluxos,omitempty"`
}

func (Quadro) TableName() string {
	return "quadros"
}

type QuadroTag struct {
	BaseModel
	QuadroID string `gorm:"not null" json:"quadroId"`
	TagID    string `gorm:"not null" json:"tagId"`

	// Relacionamentos
	Quadro Quadro `gorm:"foreignKey:QuadroID" json:"quadro,omitempty"`
	Tag    Tag    `gorm:"foreignKey:TagID" json:"tag,omitempty"`
}

func (QuadroTag) TableName() string {
	return "quadro_tags"
}

type Coluna struct {
	BaseModel
	Nome       string  `gorm:"not null" json:"nome"`
	Cor        *string `json:"cor"`
	Posicao    int     `gorm:"not null" json:"posicao"`
	QuadroID   string  `gorm:"not null" json:"quadroId"`
	AgenteIaID *string `json:"agenteIaId"`
	Ativo      bool    `gorm:"default:true" json:"ativo"`

	// Relacionamentos
	Quadro   Quadro    `gorm:"foreignKey:QuadroID" json:"quadro,omitempty"`
	AgenteIa *AgenteIa `gorm:"foreignKey:AgenteIaID" json:"agenteIa,omitempty"`
	Cards    []Card    `gorm:"foreignKey:ColunaID" json:"cards,omitempty"`
}

func (Coluna) TableName() string {
	return "colunas"
}

type Card struct {
	BaseModel
	Nome             string    `gorm:"not null" json:"nome"`
	Descricao        *string   `json:"descricao"`
	Posicao          int       `gorm:"not null" json:"posicao"`
	ColunaID         string    `gorm:"not null" json:"colunaId"`
	ConversaID       string    `gorm:"type:varchar(255);not null" json:"conversaId"`
	Prioridade       int       `gorm:"default:0" json:"prioridade"`
	DataVencimento   *string   `json:"dataVencimento"`
	Ativo            bool      `gorm:"default:true" json:"ativo"`

	// Relacionamentos
	Coluna           Coluna             `gorm:"foreignKey:ColunaID" json:"coluna,omitempty"`
	// Nota: ConversaID agora armazena IDs do WhatsApp (strings), não UUIDs do banco
	// Conversa         Conversa           `gorm:"foreignKey:ConversaID" json:"conversa,omitempty"`
	RespostasRapidas []CardRespostaRapida `gorm:"foreignKey:CardID" json:"respostasRapidas,omitempty"`
}

func (Card) TableName() string {
	return "cards"
}

// RespostaRapida foi movida para resposta_rapida.go para evitar conflitos

type CardRespostaRapida struct {
	BaseModel
	CardID           string `gorm:"not null" json:"cardId"`
	RespostaRapidaID string `gorm:"not null" json:"respostaRapidaId"`

	// Relacionamentos
	Card Card `gorm:"foreignKey:CardID" json:"card,omitempty"`
	// RespostaRapida relacionamento será definido quando necessário
}

func (CardRespostaRapida) TableName() string {
	return "card_respostas_rapidas"
}

