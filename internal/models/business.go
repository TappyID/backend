package models

import (
	"time"
)

type Atendimento struct {
	BaseModel
	Titulo       string            `gorm:"not null" json:"titulo"`
	Descricao    *string           `json:"descricao"`
	Status       StatusAtendimento `gorm:"default:AGUARDANDO" json:"status"`
	Prioridade   int               `gorm:"default:0" json:"prioridade"`
	AgenteID     *string           `json:"agenteId"`
	UsuarioID    *string           `json:"usuarioId"`
	ContatoID    string            `gorm:"not null" json:"contatoId"`
	ConversaID   string            `gorm:"not null" json:"conversaId"`
	IniciadoEm   *time.Time        `json:"iniciadoEm"`
	FinalizadoEm *time.Time        `json:"finalizadoEm"`

	// Relacionamentos
	Agente   *Usuario  `gorm:"foreignKey:AgenteID" json:"agente,omitempty"`
	Usuario  *Usuario  `gorm:"foreignKey:UsuarioID" json:"usuario,omitempty"`
	Contato  Contato   `gorm:"foreignKey:ContatoID" json:"contato,omitempty"`
	Conversa Conversa  `gorm:"foreignKey:ConversaID" json:"conversa,omitempty"`
}

func (Atendimento) TableName() string {
	return "atendimentos"
}

type Agendamento struct {
	BaseModel
	Titulo      string            `gorm:"not null" json:"titulo"`
	Descricao   *string           `json:"descricao"`
	InicioEm    time.Time         `gorm:"not null" json:"inicioEm"`
	FimEm       time.Time         `gorm:"not null" json:"fimEm"`
	LinkMeeting *string           `json:"linkMeeting"`
	Status      StatusAgendamento `gorm:"default:AGENDADO" json:"status"`
	UsuarioID   string            `gorm:"not null" json:"usuarioId"`
	ContatoID   string            `gorm:"not null" json:"contatoId"`

	// Relacionamentos
	Usuario Usuario `gorm:"foreignKey:UsuarioID" json:"usuario,omitempty"`
	Contato Contato `gorm:"foreignKey:ContatoID" json:"contato,omitempty"`
}

func (Agendamento) TableName() string {
	return "agendamentos"
}

// Enum para tipos de orçamento
type TipoOrcamento string
const (
	TipoOrcamentoVenda      TipoOrcamento = "venda"
	TipoOrcamentoAssinatura TipoOrcamento = "assinatura"
	TipoOrcamentoOrcamento  TipoOrcamento = "orcamento"
	TipoOrcamentoCobranca   TipoOrcamento = "cobranca"
)

// Enum para status de orçamento
type StatusOrcamento string
const (
	StatusOrcamentoPendente  StatusOrcamento = "PENDENTE"
	StatusOrcamentoAprovado  StatusOrcamento = "APROVADO"
	StatusOrcamentoRejeitado StatusOrcamento = "REJEITADO"
	StatusOrcamentoCancelado StatusOrcamento = "CANCELADO"
)

type Orcamento struct {
	BaseModel
	Titulo      string          `gorm:"not null" json:"titulo"`
	Data        time.Time       `gorm:"not null" json:"data"`
	Tipo        TipoOrcamento   `gorm:"not null" json:"tipo"`
	Observacao  *string         `json:"observacao"`
	ValorTotal  float64         `gorm:"not null;default:0" json:"valorTotal"`
	Status      StatusOrcamento `gorm:"default:PENDENTE" json:"status"`
	UsuarioID   string          `gorm:"not null" json:"usuarioId"`
	ContatoID   string          `gorm:"not null" json:"contatoId"`

	// Relacionamentos
	Usuario Usuario          `gorm:"foreignKey:UsuarioID" json:"usuario,omitempty"`
	Contato Contato          `gorm:"foreignKey:ContatoID" json:"contato,omitempty"`
	Itens   []OrcamentoItem  `gorm:"foreignKey:OrcamentoID" json:"itens,omitempty"`
}

func (Orcamento) TableName() string {
	return "orcamentos"
}

type OrcamentoItem struct {
	BaseModel
	Nome         string  `gorm:"not null" json:"nome"`
	Valor        float64 `gorm:"not null" json:"valor"`
	Quantidade   int     `gorm:"not null;default:1" json:"quantidade"`
	Subtotal     float64 `gorm:"not null;default:0" json:"subtotal"`
	OrcamentoID  string  `gorm:"not null" json:"orcamentoId"`

	// Relacionamentos
	Orcamento Orcamento `gorm:"foreignKey:OrcamentoID" json:"orcamento,omitempty"`
}

func (OrcamentoItem) TableName() string {
	return "orcamento_itens"
}

type Anotacao struct {
	BaseModel
	Titulo      string  `gorm:"not null" json:"titulo"`
	Conteudo    string  `gorm:"not null" json:"conteudo"`
	Importante  bool    `gorm:"default:false" json:"importante"`
	UsuarioID   string  `gorm:"not null" json:"usuarioId"`
	ContatoID   string  `gorm:"not null" json:"contatoId"`

	// Relacionamentos
	Usuario Usuario `gorm:"foreignKey:UsuarioID" json:"usuario,omitempty"`
	Contato Contato `gorm:"foreignKey:ContatoID" json:"contato,omitempty"`
}

func (Anotacao) TableName() string {
	return "anotacoes"
}

type MensagemInterna struct {
	BaseModel
	Conteudo       string `gorm:"not null" json:"conteudo"`
	RemetenteID    string `gorm:"not null" json:"remetenteId"`
	DestinatarioID string `gorm:"not null" json:"destinatarioId"`
	Lida           bool   `gorm:"default:false" json:"lida"`

	// Relacionamentos
	Remetente    Usuario `gorm:"foreignKey:RemetenteID" json:"remetente,omitempty"`
	Destinatario Usuario `gorm:"foreignKey:DestinatarioID" json:"destinatario,omitempty"`
}

func (MensagemInterna) TableName() string {
	return "mensagens_internas"
}

type AvaliacaoNps struct {
	BaseModel
	Pontuacao int     `gorm:"not null" json:"pontuacao"` // 0-10
	Feedback  *string `json:"feedback"`
	UsuarioID string  `gorm:"not null" json:"usuarioId"`
	ContatoID string  `gorm:"not null" json:"contatoId"`

	// Relacionamentos
	Usuario Usuario `gorm:"foreignKey:UsuarioID" json:"usuario,omitempty"`
	Contato Contato `gorm:"foreignKey:ContatoID" json:"contato,omitempty"`
}

func (AvaliacaoNps) TableName() string {
	return "avaliacoes_nps"
}

type Fluxo struct {
	BaseModel
	Nome       string  `gorm:"not null" json:"nome"`
	Descricao  *string `json:"descricao"`
	Ativo      bool    `gorm:"default:true" json:"ativo"`
	QuadroID   string  `gorm:"not null" json:"quadroId"`
	AgenteIaID *string `json:"agenteIaId"`

	// Relacionamentos
	Quadro   Quadro    `gorm:"foreignKey:QuadroID" json:"quadro,omitempty"`
	AgenteIa *AgenteIa `gorm:"foreignKey:AgenteIaID" json:"agenteIa,omitempty"`
	Nos      []FluxoNo `gorm:"foreignKey:FluxoID" json:"nos,omitempty"`
}

func (Fluxo) TableName() string {
	return "fluxos"
}

type FluxoNo struct {
	BaseModel
	Nome         string `gorm:"not null" json:"nome"`
	Tipo         string `gorm:"not null" json:"tipo"` // "trigger", "condition", "action", "message"
	Configuracao JSONB  `gorm:"type:jsonb" json:"configuracao"`
	Posicao      JSONB  `gorm:"type:jsonb" json:"posicao"` // Posição no canvas (x, y)
	FluxoID      string `gorm:"not null" json:"fluxoId"`

	// Relacionamentos
	Fluxo        Fluxo          `gorm:"foreignKey:FluxoID" json:"fluxo,omitempty"`
	ConexoesDe   []FluxoConexao `gorm:"foreignKey:DeID" json:"conexoesDe,omitempty"`
	ConexoesPara []FluxoConexao `gorm:"foreignKey:ParaID" json:"conexoesPara,omitempty"`
}

func (FluxoNo) TableName() string {
	return "fluxo_nos"
}

type FluxoConexao struct {
	BaseModel
	DeID   string `gorm:"not null" json:"deId"`
	ParaID string `gorm:"not null" json:"paraId"`

	// Relacionamentos
	De   FluxoNo `gorm:"foreignKey:DeID" json:"de,omitempty"`
	Para FluxoNo `gorm:"foreignKey:ParaID" json:"para,omitempty"`
}

func (FluxoConexao) TableName() string {
	return "fluxo_conexoes"
}

type Plano struct {
	BaseModel
	Nome            string  `gorm:"not null" json:"nome"`
	Descricao       *string `json:"descricao"`
	Preco           float64 `gorm:"not null" json:"preco"`
	Intervalo       string  `gorm:"not null" json:"intervalo"` // "monthly", "yearly"
	Funcionalidades JSONB   `gorm:"type:jsonb" json:"funcionalidades"`
	Ativo           bool    `gorm:"default:true" json:"ativo"`

	// Relacionamentos
	Assinaturas []Assinatura `gorm:"foreignKey:PlanoID" json:"assinaturas,omitempty"`
}

func (Plano) TableName() string {
	return "planos"
}

type Assinatura struct {
	BaseModel
	Nome                 string       `gorm:"not null" json:"nome"`
	UsuarioID            string       `gorm:"not null" json:"usuarioId"`
	ContatoID            string       `gorm:"not null" json:"contatoId"`
	PlanoID              string       `gorm:"not null" json:"planoId"`
	FormaPagamento       TipoCobranca `gorm:"not null" json:"formaPagamento"`
	LinkPagamento        *string      `json:"linkPagamento"`
	Valor                float64      `gorm:"not null" json:"valor"`
	Renovacao            string       `gorm:"not null" json:"renovacao"` // mensal, trimestral, semestral, anual, limitado
	Status               StatusPlano  `gorm:"default:ATIVO" json:"status"`
	DataInicio           time.Time    `gorm:"not null" json:"dataInicio"`
	DataFim              *time.Time   `json:"dataFim"`
	CanceladoEm          *time.Time   `json:"canceladoEm"`
	MotivoCancelamento   *string      `json:"motivoCancelamento"`

	// Relacionamentos
	Usuario   Usuario    `gorm:"foreignKey:UsuarioID" json:"usuario,omitempty"`
	Contato   Contato    `gorm:"foreignKey:ContatoID" json:"contato,omitempty"`
	Plano     Plano      `gorm:"foreignKey:PlanoID" json:"plano,omitempty"`
	Cobrancas []Cobranca `gorm:"foreignKey:AssinaturaID" json:"cobrancas,omitempty"`
}

func (Assinatura) TableName() string {
	return "assinaturas"
}

type Cobranca struct {
	BaseModel
	UsuarioID      string         `gorm:"not null" json:"usuarioId"`
	AssinaturaID   *string        `json:"assinaturaId"`
	Valor          float64        `gorm:"not null" json:"valor"`
	Tipo           TipoCobranca   `gorm:"not null" json:"tipo"`
	Status         StatusCobranca `gorm:"default:PENDENTE" json:"status"`
	DataVencimento time.Time      `gorm:"not null" json:"dataVencimento"`
	PagoEm         *time.Time     `json:"pagoEm"`
	ChavePix       *string        `json:"chavePix"`
	DadosBoleto    JSONB          `gorm:"type:jsonb" json:"dadosBoleto"`
	DadosCartao    JSONB          `gorm:"type:jsonb" json:"dadosCartao"`
	Descricao      *string        `json:"descricao"`

	// Relacionamentos
	Usuario    Usuario     `gorm:"foreignKey:UsuarioID" json:"usuario,omitempty"`
	Assinatura *Assinatura `gorm:"foreignKey:AssinaturaID" json:"assinatura,omitempty"`
}

func (Cobranca) TableName() string {
	return "cobrancas"
}

// AgenteIa representa um agente de IA para automação
type AgenteIa struct {
	ID          string     `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	UsuarioID   string     `gorm:"not null;type:uuid" json:"usuarioId"`
	Nome        string     `gorm:"not null;size:255" json:"nome"`
	Descricao   *string    `gorm:"size:1000" json:"descricao"`
	Prompt      string     `gorm:"not null;type:text" json:"prompt"`
	Modelo      string     `gorm:"not null;size:100" json:"modelo"` // deepseek, chatgpt, etc
	TokensUsados int64     `gorm:"default:0" json:"tokensUsados"`
	Nicho       *string    `gorm:"size:255" json:"nicho"`
	Ativo       bool       `gorm:"default:true" json:"ativo"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`

	// Relacionamentos
	Usuario Usuario `gorm:"foreignKey:UsuarioID" json:"usuario,omitempty"`
	Fluxos  []Fluxo `gorm:"foreignKey:AgenteIaID" json:"fluxos,omitempty"`
}

func (AgenteIa) TableName() string {
	return "agentes_ia"
}

// Contrato representa um contrato digital
type Contrato struct {
	ID           string     `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	UsuarioID    string     `gorm:"not null;type:uuid" json:"usuarioId"`
	ContatoID    *string    `gorm:"type:uuid" json:"contatoId"`
	Titulo       string     `gorm:"not null;size:255" json:"titulo"`
	Conteudo     string     `gorm:"not null;type:text" json:"conteudo"`
	Valor        *float64   `json:"valor"`
	Status       string     `gorm:"default:RASCUNHO;size:50" json:"status"` // RASCUNHO, ENVIADO, ASSINADO, CANCELADO
	DataEnvio    *time.Time `json:"dataEnvio"`
	DataAssinatura *time.Time `json:"dataAssinatura"`
	AssinaturaDigital *string `gorm:"type:text" json:"assinaturaDigital"`
	HashContrato *string    `gorm:"size:255" json:"hashContrato"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`

	// Relacionamentos
	Usuario Usuario  `gorm:"foreignKey:UsuarioID" json:"usuario,omitempty"`
	Contato *Contato `gorm:"foreignKey:ContatoID" json:"contato,omitempty"`
	Fluxos  []Fluxo  `gorm:"foreignKey:ContratoID" json:"fluxos,omitempty"`
}

func (Contrato) TableName() string {
	return "contratos"
}
