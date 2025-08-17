package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CategoriaResposta representa uma categoria de respostas rápidas
type CategoriaResposta struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Nome        string    `json:"nome" gorm:"not null;size:100"`
	Descricao   *string   `json:"descricao,omitempty"`
	Cor         string    `json:"cor" gorm:"not null;size:7;default:'#3B82F6'"`
	Icone       string    `json:"icone" gorm:"size:50;default:'Zap'"`
	UsuarioID   uuid.UUID `json:"usuario_id" gorm:"not null"`
	Ativo       bool      `json:"ativo" gorm:"default:true"`
	Ordem       int       `json:"ordem" gorm:"default:0"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relacionamentos
	Usuario          Usuario          `json:"usuario,omitempty" gorm:"foreignKey:UsuarioID"`
	RespostasRapidas []RespostaRapida `json:"respostas_rapidas,omitempty" gorm:"foreignKey:CategoriaID"`
}

// TriggerTipo define os tipos de trigger disponíveis
type TriggerTipo string

const (
	TriggerManual           TriggerTipo = "manual"
	TriggerPrimeiraMensagem TriggerTipo = "primeira_mensagem"
	TriggerPalavraChave     TriggerTipo = "palavra_chave"
	TriggerHorario          TriggerTipo = "horario"
	TriggerIntervalo        TriggerTipo = "intervalo"
)

// TriggerCondicao representa as condições de um trigger
type TriggerCondicao struct {
	PalavrasChave []string          `json:"palavras_chave,omitempty"`
	Horarios      []string          `json:"horarios,omitempty"`        // ["09:00", "18:00"]
	DiasSemanais  []int             `json:"dias_semanais,omitempty"`   // 0=domingo, 1=segunda, etc
	Intervalo     int               `json:"intervalo,omitempty"`       // Em minutos
	Condicoes     map[string]string `json:"condicoes,omitempty"`       // Condições customizadas
}

// RespostaRapida representa uma resposta rápida completa
type RespostaRapida struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Titulo      string    `json:"titulo" gorm:"not null;size:200"`
	Descricao   *string   `json:"descricao,omitempty"`
	CategoriaID uuid.UUID `json:"categoria_id" gorm:"not null"`
	UsuarioID   uuid.UUID `json:"usuario_id" gorm:"not null"`

	// Configurações de agendamento
	AgendamentoAtivo bool        `json:"agendamento_ativo" gorm:"default:false"`
	TriggerTipo      TriggerTipo `json:"trigger_tipo" gorm:"size:50;default:'manual'"`
	TriggerCondicao  *string     `json:"trigger_condicao,omitempty"` // JSON serializado

	// Configurações de execução
	DelaySegundos        int  `json:"delay_segundos" gorm:"default:0"`
	Repetir              bool `json:"repetir" gorm:"default:false"`
	IntervaloRepeticao   *int `json:"intervalo_repeticao,omitempty"` // Em minutos
	MaxRepeticoes        int  `json:"max_repeticoes" gorm:"default:1"`

	// Configurações de contato
	AplicarNovosContatos      bool    `json:"aplicar_novos_contatos" gorm:"default:true"`
	AplicarContatosExistentes bool    `json:"aplicar_contatos_existentes" gorm:"default:false"`
	ContatosEspecificos       *string `json:"contatos_especificos,omitempty"` // JSON array

	// Status
	Ativo   bool `json:"ativo" gorm:"default:true"`
	Pausado bool `json:"pausado" gorm:"default:false"`
	Ordem   int  `json:"ordem" gorm:"default:0"`

	// Estatísticas
	TotalExecucoes  int        `json:"total_execucoes" gorm:"default:0"`
	UltimaExecucao *time.Time `json:"ultima_execucao,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relacionamentos
	Categoria   CategoriaResposta `json:"categoria,omitempty" gorm:"foreignKey:CategoriaID"`
	Usuario     Usuario           `json:"usuario,omitempty" gorm:"foreignKey:UsuarioID"`
	Acoes       []AcaoResposta    `json:"acoes,omitempty" gorm:"foreignKey:RespostaRapidaID;constraint:OnDelete:CASCADE"`
	Execucoes   []ExecucaoResposta `json:"execucoes,omitempty" gorm:"foreignKey:RespostaRapidaID"`
	Agendamentos []AgendamentoResposta `json:"agendamentos,omitempty" gorm:"foreignKey:RespostaRapidaID"`
}

// TipoAcao define os tipos de ação disponíveis
type TipoAcao string

const (
	AcaoTexto       TipoAcao = "texto"
	AcaoAudio       TipoAcao = "audio"
	AcaoImagem      TipoAcao = "imagem"
	AcaoVideo       TipoAcao = "video"
	AcaoArquivo     TipoAcao = "arquivo"
	AcaoPix         TipoAcao = "pix"
	AcaoLink        TipoAcao = "link"
	AcaoContato     TipoAcao = "contato"
	AcaoLocalizacao TipoAcao = "localizacao"
	AcaoDelay       TipoAcao = "delay"
	AcaoCondicional TipoAcao = "condicional"
)

// ConteudoAcao representa o conteúdo flexível de uma ação
type ConteudoAcao map[string]interface{}

// AcaoResposta representa uma ação específica dentro de uma resposta rápida
type AcaoResposta struct {
	ID                uuid.UUID    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	RespostaRapidaID  uuid.UUID    `json:"resposta_rapida_id" gorm:"not null"`
	Tipo              TipoAcao     `json:"tipo" gorm:"not null;size:50"`
	Ordem             int          `json:"ordem" gorm:"not null;default:0"`
	DelaySegundos     int          `json:"delay_segundos" gorm:"default:0"`
	Conteudo          *string      `json:"conteudo" gorm:"type:jsonb;not null"` // JSON serializado
	Obrigatorio       bool         `json:"obrigatorio" gorm:"default:true"`
	Condicional       bool         `json:"condicional" gorm:"default:false"`
	CondicaoJSON      *string      `json:"condicao_json,omitempty"`
	Ativo             bool         `json:"ativo" gorm:"default:true"`
	CreatedAt         time.Time    `json:"created_at"`
	UpdatedAt         time.Time    `json:"updated_at"`

	// Relacionamentos
	RespostaRapida RespostaRapida `json:"resposta_rapida,omitempty" gorm:"foreignKey:RespostaRapidaID"`
}

// StatusExecucao define os status de execução
type StatusExecucao string

const (
	StatusPendente   StatusExecucao = "pendente"
	StatusExecutando StatusExecucao = "executando"
	StatusConcluida  StatusExecucao = "concluida"
	StatusErro       StatusExecucao = "erro"
	StatusCancelada  StatusExecucao = "cancelada"
)

// ExecucaoResposta representa o log de uma execução de resposta rápida
type ExecucaoResposta struct {
	ID                uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	RespostaRapidaID  uuid.UUID      `json:"resposta_rapida_id" gorm:"not null"`
	UsuarioID         uuid.UUID      `json:"usuario_id" gorm:"not null"`
	ChatID            string         `json:"chat_id" gorm:"not null;size:255"`
	ContatoNome       *string        `json:"contato_nome,omitempty" gorm:"size:255"`
	ContatoTelefone   *string        `json:"contato_telefone,omitempty" gorm:"size:50"`
	TriggerTipo       TriggerTipo    `json:"trigger_tipo" gorm:"not null;size:50"`
	TriggerDados      *string        `json:"trigger_dados,omitempty" gorm:"type:jsonb"`
	Status            StatusExecucao `json:"status" gorm:"default:'pendente';size:50"`
	AcoesExecutadas   int            `json:"acoes_executadas" gorm:"default:0"`
	TotalAcoes        int            `json:"total_acoes" gorm:"default:0"`
	ErroMensagem      *string        `json:"erro_mensagem,omitempty"`
	AgendadoPara      *time.Time     `json:"agendado_para,omitempty"`
	IniciadoEm        *time.Time     `json:"iniciado_em,omitempty"`
	ConcluidoEm       *time.Time     `json:"concluido_em,omitempty"`
	MensagensEnviadas int            `json:"mensagens_enviadas" gorm:"default:0"`
	ResultadoJSON     *string        `json:"resultado_json,omitempty" gorm:"type:jsonb"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`

	// Relacionamentos
	RespostaRapida RespostaRapida `json:"resposta_rapida,omitempty" gorm:"foreignKey:RespostaRapidaID"`
	Usuario        Usuario        `json:"usuario,omitempty" gorm:"foreignKey:UsuarioID"`
}

// AgendamentoResposta representa um agendamento ativo de resposta rápida
type AgendamentoResposta struct {
	ID                  uuid.UUID   `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	RespostaRapidaID    uuid.UUID   `json:"resposta_rapida_id" gorm:"not null"`
	UsuarioID           uuid.UUID   `json:"usuario_id" gorm:"not null"`
	ChatID              string      `json:"chat_id" gorm:"not null;size:255"`
	ContatoNome         *string     `json:"contato_nome,omitempty" gorm:"size:255"`
	ContatoTelefone     *string     `json:"contato_telefone,omitempty" gorm:"size:50"`
	TriggerTipo         TriggerTipo `json:"trigger_tipo" gorm:"not null;size:50"`
	CondicoesJSON       *string     `json:"condicoes_json,omitempty" gorm:"type:jsonb"`
	Ativo               bool        `json:"ativo" gorm:"default:true"`
	Pausado             bool        `json:"pausado" gorm:"default:false"`
	ProximaExecucao     *time.Time  `json:"proxima_execucao,omitempty"`
	ExecucoesRealizadas int         `json:"execucoes_realizadas" gorm:"default:0"`
	MaxExecucoes        int         `json:"max_execucoes" gorm:"default:1"`
	DadosContexto       *string     `json:"dados_contexto,omitempty" gorm:"type:jsonb"`
	CreatedAt           time.Time   `json:"created_at"`
	UpdatedAt           time.Time   `json:"updated_at"`

	// Relacionamentos
	RespostaRapida RespostaRapida `json:"resposta_rapida,omitempty" gorm:"foreignKey:RespostaRapidaID"`
	Usuario        Usuario        `json:"usuario,omitempty" gorm:"foreignKey:UsuarioID"`
}

// Métodos auxiliares para serialização JSON

// GetTriggerCondicao deserializa as condições do trigger
func (r *RespostaRapida) GetTriggerCondicao() (*TriggerCondicao, error) {
	if r.TriggerCondicao == nil {
		return nil, nil
	}
	
	var condicao TriggerCondicao
	err := json.Unmarshal([]byte(*r.TriggerCondicao), &condicao)
	if err != nil {
		return nil, err
	}
	
	return &condicao, nil
}

// SetTriggerCondicao serializa as condições do trigger
func (r *RespostaRapida) SetTriggerCondicao(condicao *TriggerCondicao) error {
	if condicao == nil {
		r.TriggerCondicao = nil
		return nil
	}
	
	data, err := json.Marshal(condicao)
	if err != nil {
		return err
	}
	
	condicaoStr := string(data)
	r.TriggerCondicao = &condicaoStr
	return nil
}

// GetConteudo deserializa o conteúdo da ação
func (a *AcaoResposta) GetConteudo() (ConteudoAcao, error) {
	if a.Conteudo == nil {
		return nil, nil
	}
	
	var conteudo ConteudoAcao
	err := json.Unmarshal([]byte(*a.Conteudo), &conteudo)
	if err != nil {
		return nil, err
	}
	
	return conteudo, nil
}

// SetConteudo serializa o conteúdo da ação
func (a *AcaoResposta) SetConteudo(conteudo ConteudoAcao) error {
	if conteudo == nil {
		a.Conteudo = nil
		return nil
	}
	
	data, err := json.Marshal(conteudo)
	if err != nil {
		return err
	}
	
	conteudoStr := string(data)
	a.Conteudo = &conteudoStr
	return nil
}

// GetContatosEspecificos deserializa a lista de contatos específicos
func (r *RespostaRapida) GetContatosEspecificos() ([]string, error) {
	if r.ContatosEspecificos == nil {
		return nil, nil
	}
	
	var contatos []string
	err := json.Unmarshal([]byte(*r.ContatosEspecificos), &contatos)
	if err != nil {
		return nil, err
	}
	
	return contatos, nil
}

// SetContatosEspecificos serializa a lista de contatos específicos
func (r *RespostaRapida) SetContatosEspecificos(contatos []string) error {
	if contatos == nil || len(contatos) == 0 {
		r.ContatosEspecificos = nil
		return nil
	}
	
	data, err := json.Marshal(contatos)
	if err != nil {
		return err
	}
	
	contatosStr := string(data)
	r.ContatosEspecificos = &contatosStr
	return nil
}

// Hooks do GORM
func (r *RespostaRapida) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

func (c *CategoriaResposta) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

func (a *AcaoResposta) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

func (e *ExecucaoResposta) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

func (a *AgendamentoResposta) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}
