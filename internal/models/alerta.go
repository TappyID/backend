package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type Alerta struct {
	BaseModel
	Titulo    string           `gorm:"not null" json:"titulo"`
	Descricao string           `gorm:"type:text" json:"descricao"`
	Tipo      TipoAlerta       `gorm:"not null" json:"tipo"`
	Prioridade PrioridadeAlerta `gorm:"not null" json:"prioridade"`
	Status    StatusAlerta     `gorm:"not null;default:'ativo'" json:"status"`
	Cor       string           `gorm:"not null" json:"cor"`
	Icone     string           `gorm:"not null" json:"icone"`
	Configuracoes ConfiguracaoAlerta `gorm:"type:jsonb" json:"configuracoes"`
	Estatisticas  EstatisticasAlerta `gorm:"type:jsonb" json:"estatisticas"`

	// Relacionamentos
	UserID string  `gorm:"column:user_id;not null" json:"userId"`
	Usuario Usuario `gorm:"foreignKey:UserID" json:"usuario,omitempty"`
}

func (Alerta) TableName() string {
	return "alertas"
}

type TipoAlerta string

const (
	TipoAlertaSistema     TipoAlerta = "sistema"
	TipoAlertaUsuario     TipoAlerta = "usuario"
	TipoAlertaSeguranca   TipoAlerta = "seguranca"
	TipoAlertaPerformance TipoAlerta = "performance"
	TipoAlertaIntegracao  TipoAlerta = "integracao"
)

type PrioridadeAlerta string

const (
	PrioridadeAlertaBaixa   PrioridadeAlerta = "baixa"
	PrioridadeAlertaMedia   PrioridadeAlerta = "media"
	PrioridadeAlertaAlta    PrioridadeAlerta = "alta"
	PrioridadeAlertaCritica PrioridadeAlerta = "critica"
)

type StatusAlerta string

const (
	StatusAlertaAtivo     StatusAlerta = "ativo"
	StatusAlertaPausado   StatusAlerta = "pausado"
	StatusAlertaResolvido StatusAlerta = "resolvido"
)

type ConfiguracaoAlerta struct {
	EmailNotificacao     bool              `json:"emailNotificacao"`
	WhatsappNotificacao  bool              `json:"whatsappNotificacao"`
	DashboardNotificacao bool              `json:"dashboardNotificacao"`
	Frequencia           FrequenciaAlerta  `json:"frequencia"`
	Destinatarios        []string          `json:"destinatarios"`
	Condicoes            []CondicaoAlerta  `json:"condicoes"`
}

type FrequenciaAlerta string

const (
	FrequenciaAlertaImediata FrequenciaAlerta = "imediata"
	FrequenciaAlertaHoraria  FrequenciaAlerta = "horaria"
	FrequenciaAlertaDiaria   FrequenciaAlerta = "diaria"
	FrequenciaAlertaSemanal  FrequenciaAlerta = "semanal"
)

type CondicaoAlerta struct {
	Metrica   string      `json:"metrica"`
	Operador  string      `json:"operador"` // >, <, =, >=, <=, !=
	Valor     interface{} `json:"valor"`
}

type EstatisticasAlerta struct {
	TotalDisparos   int    `json:"totalDisparos"`
	DisparosHoje    int    `json:"disparosHoje"`
	UltimoDisparo   *BaseModel `json:"ultimoDisparo,omitempty"`
	TaxaResolucao   float64 `json:"taxaResolucao"`
}

// Implementar driver.Valuer e sql.Scanner para ConfiguracaoAlerta
func (c ConfiguracaoAlerta) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *ConfiguracaoAlerta) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, c)
}

// Implementar driver.Valuer e sql.Scanner para EstatisticasAlerta
func (e EstatisticasAlerta) Value() (driver.Value, error) {
	return json.Marshal(e)
}

func (e *EstatisticasAlerta) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, e)
}

type HistoricoAlerta struct {
	BaseModel
	AlertaID  string      `gorm:"not null" json:"alertaId"`
	Alerta    Alerta      `gorm:"foreignKey:AlertaID" json:"alerta,omitempty"`
	Disparado bool        `gorm:"not null" json:"disparado"`
	Mensagem  string      `gorm:"type:text" json:"mensagem"`
	Dados     interface{} `gorm:"type:jsonb" json:"dados,omitempty"`
}

func (HistoricoAlerta) TableName() string {
	return "historico_alertas"
}
