package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Base model com campos comuns
type BaseModel struct {
	ID           string    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CriadoEm     time.Time `gorm:"autoCreateTime" json:"criadoEm"`
	AtualizadoEm time.Time `gorm:"autoUpdateTime" json:"atualizadoEm"`
}

// BeforeCreate hook para gerar UUID
func (base *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if base.ID == "" {
		base.ID = uuid.New().String()
	}
	return nil
}

// Enums
type TipoUsuario string
const (
	TipoUsuarioAdmin               TipoUsuario = "ADMIN"
	TipoUsuarioAtendenteFinanceiro TipoUsuario = "ATENDENTE_FINANCEIRO"
	TipoUsuarioAtendenteComercial  TipoUsuario = "ATENDENTE_COMERCIAL"
	TipoUsuarioAtendenteJuridico   TipoUsuario = "ATENDENTE_JURIDICO"
	TipoUsuarioAtendenteSuporte    TipoUsuario = "ATENDENTE_SUPORTE"
	TipoUsuarioAtendenteVendas     TipoUsuario = "ATENDENTE_VENDAS"
	TipoUsuarioAssinante           TipoUsuario = "ASSINANTE"
	TipoUsuarioAfiliado            TipoUsuario = "AFILIADO"
)

type StatusSessao string
const (
	StatusSessaoDesconectado StatusSessao = "DESCONECTADO"
	StatusSessaoConectando   StatusSessao = "CONECTANDO"
	StatusSessaoConectado    StatusSessao = "CONECTADO"
	StatusSessaoAutenticado  StatusSessao = "AUTENTICADO"
	StatusSessaoFalhou       StatusSessao = "FALHOU"
)

type TipoMensagem string
const (
	TipoMensagemTexto         TipoMensagem = "TEXTO"
	TipoMensagemImagem        TipoMensagem = "IMAGEM"
	TipoMensagemArquivo       TipoMensagem = "ARQUIVO"
	TipoMensagemAudio         TipoMensagem = "AUDIO"
	TipoMensagemVideo         TipoMensagem = "VIDEO"
	TipoMensagemLocalizacao   TipoMensagem = "LOCALIZACAO"
	TipoMensagemContato       TipoMensagem = "CONTATO"
	TipoMensagemEnquete       TipoMensagem = "ENQUETE"
	TipoMensagemRespostaBotao TipoMensagem = "RESPOSTA_BOTAO"
)

type StatusMensagem string
const (
	StatusMensagemPendente  StatusMensagem = "PENDENTE"
	StatusMensagemEnviado   StatusMensagem = "ENVIADO"
	StatusMensagemEntregue  StatusMensagem = "ENTREGUE"
	StatusMensagemLido      StatusMensagem = "LIDO"
	StatusMensagemFalhou    StatusMensagem = "FALHOU"
)

type StatusAtendimento string
const (
	StatusAtendimentoAguardando  StatusAtendimento = "AGUARDANDO"
	StatusAtendimentoEmAndamento StatusAtendimento = "EM_ANDAMENTO"
	StatusAtendimentoFinalizado  StatusAtendimento = "FINALIZADO"
	StatusAtendimentoCancelado   StatusAtendimento = "CANCELADO"
)

type StatusAgendamento string
const (
	StatusAgendamentoAgendado    StatusAgendamento = "AGENDADO"
	StatusAgendamentoConfirmado  StatusAgendamento = "CONFIRMADO"
	StatusAgendamentoRealizado   StatusAgendamento = "REALIZADO"
	StatusAgendamentoCancelado   StatusAgendamento = "CANCELADO"
	StatusAgendamentoReagendado  StatusAgendamento = "REAGENDADO"
)

type StatusPlano string
const (
	StatusPlanoAtivo     StatusPlano = "ATIVO"
	StatusPlanoInativo   StatusPlano = "INATIVO"
	StatusPlanoCancelado StatusPlano = "CANCELADO"
	StatusPlanoSuspenso  StatusPlano = "SUSPENSO"
)

type StatusCobranca string
const (
	StatusCobrancaPendente  StatusCobranca = "PENDENTE"
	StatusCobrancaPago      StatusCobranca = "PAGO"
	StatusCobrancaVencido   StatusCobranca = "VENCIDO"
	StatusCobrancaCancelado StatusCobranca = "CANCELADO"
)

type TipoCobranca string
const (
	TipoCobrancaPix    TipoCobranca = "PIX"
	TipoCobrancaCartao TipoCobranca = "CARTAO"
	TipoCobrancaBoleto TipoCobranca = "BOLETO"
)

// Models
type Usuario struct {
	BaseModel
	Email     string      `gorm:"uniqueIndex;not null" json:"email"`
	Nome      string      `gorm:"not null" json:"nome"`
	Telefone  *string     `json:"telefone"`
	Avatar    *string     `json:"avatar"`
	Tipo      TipoUsuario `gorm:"not null" json:"tipo"`
	Ativo     bool        `gorm:"default:true" json:"ativo"`
	Senha     string      `gorm:"not null" json:"-"` // Não retornar na API

	// Relacionamentos
	Sessoes             []SessaoWhatsApp  `gorm:"foreignKey:UsuarioID" json:"sessoes,omitempty"`
	AtendimentosAgente  []Atendimento     `gorm:"foreignKey:AgenteID" json:"atendimentosAgente,omitempty"`
	AtendimentosUsuario []Atendimento     `gorm:"foreignKey:UsuarioID" json:"atendimentosUsuario,omitempty"`
	Quadros             []Quadro          `gorm:"foreignKey:UsuarioID" json:"quadros,omitempty"`
	RespostasRapidas    []RespostaRapida  `gorm:"foreignKey:UsuarioID" json:"respostasRapidas,omitempty"`
	Agendamentos        []Agendamento     `gorm:"foreignKey:UsuarioID" json:"agendamentos,omitempty"`
	Orcamentos          []Orcamento       `gorm:"foreignKey:UsuarioID" json:"orcamentos,omitempty"`
	Anotacoes           []Anotacao        `gorm:"foreignKey:UsuarioID" json:"anotacoes,omitempty"`
	MensagensEnviadas   []MensagemInterna `gorm:"foreignKey:RemetenteID" json:"mensagensEnviadas,omitempty"`
	MensagensRecebidas  []MensagemInterna `gorm:"foreignKey:DestinatarioID" json:"mensagensRecebidas,omitempty"`
	AvaliacoesNps       []AvaliacaoNps    `gorm:"foreignKey:UsuarioID" json:"avaliacoesNps,omitempty"`
	Assinaturas         []Assinatura      `gorm:"foreignKey:UsuarioID" json:"assinaturas,omitempty"`
	Cobrancas           []Cobranca        `gorm:"foreignKey:UsuarioID" json:"cobrancas,omitempty"`
}

func (Usuario) TableName() string {
	return "usuarios"
}

type SessaoWhatsApp struct {
	BaseModel
	NomeSessao      string       `gorm:"uniqueIndex;not null" json:"nomeSessao"`
	NumeroTelefone  *string      `json:"numeroTelefone"`
	Status          StatusSessao `gorm:"default:DESCONECTADO" json:"status"`
	CodigoQr        *string      `json:"codigoQr"`
	UrlWebhook      *string      `json:"urlWebhook"`
	Ativo           bool         `gorm:"default:true" json:"ativo"`
	UsuarioID       string       `gorm:"not null" json:"usuarioId"`

	// Relacionamentos
	Usuario   Usuario    `gorm:"foreignKey:UsuarioID" json:"usuario,omitempty"`
	Contatos  []Contato  `gorm:"foreignKey:SessaoWhatsappID" json:"contatos,omitempty"`
	Conversas []Conversa `gorm:"foreignKey:SessaoWhatsappID" json:"conversas,omitempty"`
}

func (SessaoWhatsApp) TableName() string {
	return "sessoes_whatsapp"
}

type Contato struct {
	BaseModel
	NumeroTelefone   string  `gorm:"not null" json:"numeroTelefone"`
	Nome             *string `json:"nome"`
	FotoPerfil       *string `json:"fotoPerfil"`
	Sobre            *string `json:"sobre"`
	Bloqueado        bool    `gorm:"default:false" json:"bloqueado"`
	SessaoWhatsappID string  `gorm:"not null" json:"sessaoWhatsappId"`
	
	// Campos adicionais do contato
	Email    *string `json:"email"`
	Empresa  *string `json:"empresa"`
	CPF      *string `json:"cpf"`
	CNPJ     *string `json:"cnpj"`
	
	// Campos de endereço
	CEP     *string `json:"cep"`
	Rua     *string `json:"rua"`
	Numero  *string `json:"numero"`
	Bairro  *string `json:"bairro"`
	Cidade  *string `json:"cidade"`
	Estado  *string `json:"estado"`
	Pais    *string `json:"pais"`

	// Campos Kanban
	UltimaInteracao *time.Time `json:"ultimaInteracao,omitempty"`
	StatusKanban    *string    `json:"statusKanban,omitempty"`
	Favorito        bool       `gorm:"default:false" json:"favorito"`

	// Relacionamentos
	SessaoWhatsapp SessaoWhatsApp  `gorm:"foreignKey:SessaoWhatsappID" json:"sessaoWhatsapp,omitempty"`
	Conversas      []Conversa      `gorm:"foreignKey:ContatoID" json:"conversas,omitempty"`
	Tags           []ContatoTag    `gorm:"foreignKey:ContatoID" json:"tags,omitempty"`
	Atendimentos   []Atendimento   `gorm:"foreignKey:ContatoID" json:"atendimentos,omitempty"`
	Agendamentos   []Agendamento   `gorm:"foreignKey:ContatoID" json:"agendamentos,omitempty"`
	Orcamentos     []Orcamento     `gorm:"foreignKey:ContatoID" json:"orcamentos,omitempty"`
	Assinaturas    []Assinatura    `gorm:"foreignKey:ContatoID" json:"assinaturas,omitempty"`
	Anotacoes      []Anotacao      `gorm:"foreignKey:ContatoID" json:"anotacoes,omitempty"`
	AvaliacoesNps  []AvaliacaoNps  `gorm:"foreignKey:ContatoID" json:"avaliacoesNps,omitempty"`
}

func (Contato) TableName() string {
	return "contatos"
}

type Conversa struct {
	BaseModel
	IDConversa               string     `gorm:"not null" json:"idConversa"`
	Nome                     *string    `json:"nome"`
	EhGrupo                  bool       `gorm:"default:false" json:"ehGrupo"`
	FotoPerfil               *string    `json:"fotoPerfil"`
	UltimaMensagem           *string    `json:"ultimaMensagem"`
	HorarioUltimaMensagem    *time.Time `json:"horarioUltimaMensagem"`
	MensagensNaoLidas        int        `gorm:"default:0" json:"mensagensNaoLidas"`
	Arquivada                bool       `gorm:"default:false" json:"arquivada"`
	SessaoWhatsappID         string     `gorm:"not null" json:"sessaoWhatsappId"`
	ContatoID                *string    `json:"contatoId"`

	// Relacionamentos
	SessaoWhatsapp SessaoWhatsApp `gorm:"foreignKey:SessaoWhatsappID" json:"sessaoWhatsapp,omitempty"`
	Contato        *Contato       `gorm:"foreignKey:ContatoID" json:"contato,omitempty"`
	Mensagens      []Mensagem     `gorm:"foreignKey:ConversaID" json:"mensagens,omitempty"`
	Atendimentos   []Atendimento  `gorm:"foreignKey:ConversaID" json:"atendimentos,omitempty"`
	Cards          []Card         `gorm:"foreignKey:ConversaID" json:"cards,omitempty"`
}

func (Conversa) TableName() string {
	return "conversas"
}

type Mensagem struct {
	BaseModel
	IDMensagem     string         `gorm:"not null" json:"idMensagem"`
	ConversaID     string         `gorm:"not null" json:"conversaId"`
	DeMim          bool           `json:"deMim"`
	Tipo           TipoMensagem   `json:"tipo"`
	Conteudo       *string        `json:"conteudo"`
	UrlMidia       *string        `json:"urlMidia"`
	Legenda        *string        `json:"legenda"`
	Status         StatusMensagem `gorm:"default:PENDENTE" json:"status"`
	Timestamp      time.Time      `json:"timestamp"`
	RespostaParaID *string        `json:"respostaParaId"`
	Encaminhada    bool           `gorm:"default:false" json:"encaminhada"`
	Favorita       bool           `gorm:"default:false" json:"favorita"`

	// Relacionamentos
	Conversa     Conversa   `gorm:"foreignKey:ConversaID" json:"conversa,omitempty"`
	RespostaPara *Mensagem  `gorm:"foreignKey:RespostaParaID" json:"respostaPara,omitempty"`
	Respostas    []Mensagem `gorm:"foreignKey:RespostaParaID" json:"respostas,omitempty"`
}

func (Mensagem) TableName() string {
	return "mensagens"
}
