package services

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"tappyone/internal/models"
	"tappyone/internal/repositories"
)

type RespostaRapidaService struct {
	repo            *repositories.RespostaRapidaRepository
	whatsappService *WhatsAppService
}

func NewRespostaRapidaService(repo *repositories.RespostaRapidaRepository, whatsappService *WhatsAppService) *RespostaRapidaService {
	return &RespostaRapidaService{
		repo:            repo,
		whatsappService: whatsappService,
	}
}

// ===== CATEGORIAS =====

func (s *RespostaRapidaService) CreateCategoria(usuarioID uuid.UUID, nome, descricao, cor, icone string) (*models.CategoriaResposta, error) {
	categoria := &models.CategoriaResposta{
		Nome:      nome,
		Descricao: &descricao,
		Cor:       cor,
		Icone:     icone,
		UsuarioID: usuarioID,
		Ativo:     true,
		Ordem:     0,
	}

	err := s.repo.CreateCategoria(categoria)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar categoria: %w", err)
	}

	return categoria, nil
}

func (s *RespostaRapidaService) GetCategoriasByUsuario(usuarioID uuid.UUID) ([]models.CategoriaResposta, error) {
	return s.repo.GetCategoriasByUsuario(usuarioID)
}

func (s *RespostaRapidaService) UpdateCategoria(id uuid.UUID, nome, descricao, cor, icone string) error {
	categoria, err := s.repo.GetCategoriaByID(id)
	if err != nil {
		return fmt.Errorf("categoria não encontrada: %w", err)
	}

	categoria.Nome = nome
	if descricao != "" {
		categoria.Descricao = &descricao
	}
	categoria.Cor = cor
	categoria.Icone = icone

	return s.repo.UpdateCategoria(categoria)
}

func (s *RespostaRapidaService) DeleteCategoria(id uuid.UUID) error {
	return s.repo.DeleteCategoria(id)
}

// ===== RESPOSTAS RÁPIDAS =====

func (s *RespostaRapidaService) CreateRespostaRapida(req *CreateRespostaRapidaRequest) (*models.RespostaRapida, error) {
	resposta := &models.RespostaRapida{
		Titulo:                    req.Titulo,
		Descricao:                 req.Descricao,
		CategoriaID:               req.CategoriaID,
		UsuarioID:                 req.UsuarioID,
		AgendamentoAtivo:          req.AgendamentoAtivo,
		TriggerTipo:               req.TriggerTipo,
		DelaySegundos:             req.DelaySegundos,
		Repetir:                   req.Repetir,
		IntervaloRepeticao:        req.IntervaloRepeticao,
		MaxRepeticoes:             req.MaxRepeticoes,
		AplicarNovosContatos:      req.AplicarNovosContatos,
		AplicarContatosExistentes: req.AplicarContatosExistentes,
		Ativo:                     true,
		Pausado:                   false,
		Ordem:                     0,
	}

	// Serializar condições do trigger
	if req.TriggerCondicao != nil {
		err := resposta.SetTriggerCondicao(req.TriggerCondicao)
		if err != nil {
			return nil, fmt.Errorf("erro ao serializar condições do trigger: %w", err)
		}
	}

	// Serializar contatos específicos
	if req.ContatosEspecificos != nil && len(req.ContatosEspecificos) > 0 {
		err := resposta.SetContatosEspecificos(req.ContatosEspecificos)
		if err != nil {
			return nil, fmt.Errorf("erro ao serializar contatos específicos: %w", err)
		}
	}

	err := s.repo.CreateRespostaRapida(resposta)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar resposta rápida: %w", err)
	}

	// Criar ações se fornecidas
	if req.Acoes != nil && len(req.Acoes) > 0 {
		for i, acaoReq := range req.Acoes {
			acao := &models.AcaoResposta{
				RespostaRapidaID: resposta.ID,
				Tipo:             acaoReq.Tipo,
				Ordem:            i,
				DelaySegundos:    acaoReq.DelaySegundos,
				Obrigatorio:      acaoReq.Obrigatorio,
				Condicional:      acaoReq.Condicional,
				Ativo:            true,
			}

			err := acao.SetConteudo(acaoReq.Conteudo)
			if err != nil {
				return nil, fmt.Errorf("erro ao serializar conteúdo da ação: %w", err)
			}

			err = s.repo.CreateAcao(acao)
			if err != nil {
				return nil, fmt.Errorf("erro ao criar ação: %w", err)
			}
		}
	}

	// Buscar resposta completa com relacionamentos
	return s.repo.GetRespostaRapidaByID(resposta.ID)
}

func (s *RespostaRapidaService) GetRespostasRapidasByUsuario(usuarioID uuid.UUID) ([]models.RespostaRapida, error) {
	return s.repo.GetRespostasRapidasByUsuario(usuarioID)
}

func (s *RespostaRapidaService) GetRespostaRapidaByID(id uuid.UUID) (*models.RespostaRapida, error) {
	return s.repo.GetRespostaRapidaByID(id)
}

func (s *RespostaRapidaService) TogglePausarRespostaRapida(id uuid.UUID, pausado bool) error {
	return s.repo.TogglePausarRespostaRapida(id, pausado)
}

// ===== EXECUÇÃO DE RESPOSTAS RÁPIDAS =====

// ExecutarRespostaRapida executa uma resposta rápida manualmente
func (s *RespostaRapidaService) ExecutarRespostaRapida(respostaID uuid.UUID, chatID string, usuarioID uuid.UUID) error {
	resposta, err := s.repo.GetRespostaRapidaByID(respostaID)
	if err != nil {
		return fmt.Errorf("resposta rápida não encontrada: %w", err)
	}

	if !resposta.Ativo || resposta.Pausado {
		return fmt.Errorf("resposta rápida não está ativa")
	}

	// Criar execução
	execucao := &models.ExecucaoResposta{
		RespostaRapidaID: respostaID,
		UsuarioID:        usuarioID,
		ChatID:           chatID,
		TriggerTipo:      models.TriggerManual,
		Status:           models.StatusPendente,
		TotalAcoes:       len(resposta.Acoes),
	}

	err = s.repo.CreateExecucao(execucao)
	if err != nil {
		return fmt.Errorf("erro ao criar execução: %w", err)
	}

	// Executar ações
	go s.processarExecucao(execucao.ID)

	return nil
}

// ProcessarMensagemRecebida processa uma mensagem recebida e verifica se deve disparar respostas rápidas
func (s *RespostaRapidaService) ProcessarMensagemRecebida(usuarioID uuid.UUID, chatID, mensagem, contatoNome, contatoTelefone string) error {
	// Buscar respostas rápidas ativas do usuário
	respostas, err := s.repo.GetRespostasRapidasAtivas(usuarioID)
	if err != nil {
		return fmt.Errorf("erro ao buscar respostas ativas: %w", err)
	}

	for _, resposta := range respostas {
		// Verificar se deve disparar esta resposta
		shouldTrigger, err := s.shouldTriggerResposta(&resposta, chatID, mensagem, usuarioID)
		if err != nil {
			log.Printf("Erro ao verificar trigger para resposta %s: %v", resposta.ID, err)
			continue
		}

		if shouldTrigger {
			// Criar execução
			execucao := &models.ExecucaoResposta{
				RespostaRapidaID: resposta.ID,
				UsuarioID:        usuarioID,
				ChatID:           chatID,
				ContatoNome:      &contatoNome,
				ContatoTelefone:  &contatoTelefone,
				TriggerTipo:      resposta.TriggerTipo,
				Status:           models.StatusPendente,
				TotalAcoes:       len(resposta.Acoes),
			}

			// Aplicar delay se configurado
			if resposta.DelaySegundos > 0 {
				agendadoPara := time.Now().Add(time.Duration(resposta.DelaySegundos) * time.Second)
				execucao.AgendadoPara = &agendadoPara
			}

			err = s.repo.CreateExecucao(execucao)
			if err != nil {
				log.Printf("Erro ao criar execução para resposta %s: %v", resposta.ID, err)
				continue
			}

			// Criar/atualizar agendamento se necessário
			if resposta.Repetir {
				err = s.criarOuAtualizarAgendamento(&resposta, chatID, contatoNome, contatoTelefone, usuarioID)
				if err != nil {
					log.Printf("Erro ao criar agendamento para resposta %s: %v", resposta.ID, err)
				}
			}

			// Executar imediatamente se não há delay
			if resposta.DelaySegundos == 0 {
				go s.processarExecucao(execucao.ID)
			}
		}
	}

	return nil
}

// shouldTriggerResposta verifica se uma resposta deve ser disparada
func (s *RespostaRapidaService) shouldTriggerResposta(resposta *models.RespostaRapida, chatID, mensagem string, usuarioID uuid.UUID) (bool, error) {
	switch resposta.TriggerTipo {
	case models.TriggerPrimeiraMensagem:
		// Verificar se é a primeira mensagem do contato
		// TODO: Implementar lógica para verificar histórico de mensagens
		return true, nil

	case models.TriggerPalavraChave:
		condicao, err := resposta.GetTriggerCondicao()
		if err != nil {
			return false, err
		}
		if condicao != nil && condicao.PalavrasChave != nil {
			mensagemLower := strings.ToLower(mensagem)
			for _, palavra := range condicao.PalavrasChave {
				if strings.Contains(mensagemLower, strings.ToLower(palavra)) {
					return true, nil
				}
			}
		}
		return false, nil

	case models.TriggerHorario:
		condicao, err := resposta.GetTriggerCondicao()
		if err != nil {
			return false, err
		}
		if condicao != nil && condicao.Horarios != nil {
			agora := time.Now()
			horaAtual := agora.Format("15:04")
			diaAtual := int(agora.Weekday())

			// Verificar horário
			for _, horario := range condicao.Horarios {
				if horario == horaAtual {
					// Verificar dia da semana se especificado
					if condicao.DiasSemanais == nil || len(condicao.DiasSemanais) == 0 {
						return true, nil
					}
					for _, dia := range condicao.DiasSemanais {
						if dia == diaAtual {
							return true, nil
						}
					}
				}
			}
		}
		return false, nil

	default:
		return false, nil
	}
}

// criarOuAtualizarAgendamento cria ou atualiza um agendamento para repetição
func (s *RespostaRapidaService) criarOuAtualizarAgendamento(resposta *models.RespostaRapida, chatID, contatoNome, contatoTelefone string, usuarioID uuid.UUID) error {
	// Verificar se já existe agendamento
	agendamento, err := s.repo.GetAgendamentoByChat(resposta.ID, chatID, usuarioID)
	if err != nil {
		// Criar novo agendamento
		agendamento = &models.AgendamentoResposta{
			RespostaRapidaID:    resposta.ID,
			UsuarioID:           usuarioID,
			ChatID:              chatID,
			ContatoNome:         &contatoNome,
			ContatoTelefone:     &contatoTelefone,
			TriggerTipo:         resposta.TriggerTipo,
			Ativo:               true,
			Pausado:             false,
			ExecucoesRealizadas: 1,
			MaxExecucoes:        resposta.MaxRepeticoes,
		}

		// Calcular próxima execução
		if resposta.IntervaloRepeticao != nil {
			proximaExecucao := time.Now().Add(time.Duration(*resposta.IntervaloRepeticao) * time.Minute)
			agendamento.ProximaExecucao = &proximaExecucao
		}

		return s.repo.CreateAgendamento(agendamento)
	} else {
		// Atualizar agendamento existente
		agendamento.ExecucoesRealizadas++
		
		// Verificar se ainda deve continuar repetindo
		if agendamento.ExecucoesRealizadas >= agendamento.MaxExecucoes {
			agendamento.Ativo = false
		} else if resposta.IntervaloRepeticao != nil {
			proximaExecucao := time.Now().Add(time.Duration(*resposta.IntervaloRepeticao) * time.Minute)
			agendamento.ProximaExecucao = &proximaExecucao
		}

		return s.repo.UpdateAgendamento(agendamento)
	}
}

// processarExecucao processa uma execução de resposta rápida
func (s *RespostaRapidaService) processarExecucao(execucaoID uuid.UUID) {
	// Buscar execução
	execucoes, err := s.repo.GetExecucoesPendentes()
	if err != nil {
		log.Printf("Erro ao buscar execuções pendentes: %v", err)
		return
	}

	var execucao *models.ExecucaoResposta
	for _, e := range execucoes {
		if e.ID == execucaoID {
			execucao = &e
			break
		}
	}

	if execucao == nil {
		log.Printf("Execução %s não encontrada", execucaoID)
		return
	}

	// Atualizar status para executando
	agora := time.Now()
	execucao.Status = models.StatusExecutando
	execucao.IniciadoEm = &agora

	err = s.repo.UpdateExecucao(execucao)
	if err != nil {
		log.Printf("Erro ao atualizar status da execução: %v", err)
		return
	}

	// Executar ações
	for i, acao := range execucao.RespostaRapida.Acoes {
		if !acao.Ativo {
			continue
		}

		// Aplicar delay da ação
		if acao.DelaySegundos > 0 {
			time.Sleep(time.Duration(acao.DelaySegundos) * time.Second)
		}

		err := s.executarAcao(&acao, execucao.ChatID, execucao.UsuarioID)
		if err != nil {
			log.Printf("Erro ao executar ação %s: %v", acao.ID, err)
			
			if acao.Obrigatorio {
				// Se a ação é obrigatória e falhou, parar execução
				execucao.Status = models.StatusErro
				errMsg := err.Error()
				execucao.ErroMensagem = &errMsg
				execucao.AcoesExecutadas = i
				break
			}
		} else {
			execucao.AcoesExecutadas = i + 1
			execucao.MensagensEnviadas++
		}
	}

	// Finalizar execução
	if execucao.Status != models.StatusErro {
		execucao.Status = models.StatusConcluida
	}
	
	agora = time.Now()
	execucao.ConcluidoEm = &agora

	err = s.repo.UpdateExecucao(execucao)
	if err != nil {
		log.Printf("Erro ao finalizar execução: %v", err)
	}

	// Atualizar estatísticas da resposta rápida
	resposta := execucao.RespostaRapida
	resposta.TotalExecucoes++
	resposta.UltimaExecucao = &agora
	s.repo.UpdateRespostaRapida(&resposta)
}

// executarAcao executa uma ação específica
func (s *RespostaRapidaService) executarAcao(acao *models.AcaoResposta, chatID string, usuarioID uuid.UUID) error {
	conteudo, err := acao.GetConteudo()
	if err != nil {
		return fmt.Errorf("erro ao deserializar conteúdo da ação: %w", err)
	}

	// Converter usuarioID para sessionName
	sessionName := fmt.Sprintf("user_%s", usuarioID.String())

	switch acao.Tipo {
	case models.AcaoTexto:
		mensagem, ok := conteudo["mensagem"].(string)
		if !ok {
			return fmt.Errorf("mensagem não encontrada no conteúdo da ação")
		}
		
		// TODO: Processar variáveis se necessário
		mensagem = s.processarVariaveis(mensagem, chatID, usuarioID)
		
		_, err := s.whatsappService.SendMessage(sessionName, chatID, mensagem)
		return err

	case models.AcaoImagem:
		arquivoURL, ok := conteudo["arquivo_url"].(string)
		if !ok {
			return fmt.Errorf("arquivo_url não encontrado no conteúdo da ação")
		}
		
		legenda, _ := conteudo["legenda"].(string)
		
		_, err := s.whatsappService.SendImage(sessionName, chatID, arquivoURL, legenda)
		return err

	case models.AcaoAudio:
		tipoAudio, ok := conteudo["tipo"].(string)
		if !ok {
			return fmt.Errorf("tipo de áudio não especificado")
		}
		
		if tipoAudio == "arquivo" {
			arquivoURL, ok := conteudo["arquivo_url"].(string)
			if !ok {
				return fmt.Errorf("arquivo_url não encontrado para áudio")
			}
			_, err := s.whatsappService.SendVoice(sessionName, chatID, arquivoURL)
			return err
		}
		
		// TODO: Implementar geração de áudio com IA
		return fmt.Errorf("geração de áudio com IA não implementada ainda")

	case models.AcaoVideo:
		arquivoURL, ok := conteudo["arquivo_url"].(string)
		if !ok {
			return fmt.Errorf("arquivo_url não encontrado no conteúdo da ação")
		}
		
		legenda, _ := conteudo["legenda"].(string)
		
		_, err := s.whatsappService.SendVideo(sessionName, chatID, arquivoURL, legenda)
		return err

	case models.AcaoArquivo:
		arquivoURL, ok := conteudo["arquivo_url"].(string)
		if !ok {
			return fmt.Errorf("arquivo_url não encontrado no conteúdo da ação")
		}
		
		filename, _ := conteudo["filename"].(string)
		legenda, _ := conteudo["legenda"].(string)
		
		_, err := s.whatsappService.SendFile(sessionName, chatID, arquivoURL, filename, legenda)
		return err

	case models.AcaoPix:
		// TODO: Implementar geração de PIX
		return fmt.Errorf("geração de PIX não implementada ainda")

	case models.AcaoDelay:
		segundos, ok := conteudo["segundos"].(float64)
		if !ok {
			return fmt.Errorf("segundos não especificados para delay")
		}
		
		time.Sleep(time.Duration(segundos) * time.Second)
		return nil

	default:
		return fmt.Errorf("tipo de ação não suportado: %s", acao.Tipo)
	}
}

// processarVariaveis processa variáveis na mensagem
func (s *RespostaRapidaService) processarVariaveis(mensagem, chatID string, usuarioID uuid.UUID) string {
	// TODO: Implementar processamento de variáveis
	// Exemplos: {nome_cliente}, {horario_atual}, {data_atual}, etc.
	
	agora := time.Now()
	mensagem = strings.ReplaceAll(mensagem, "{horario_atual}", agora.Format("15:04"))
	mensagem = strings.ReplaceAll(mensagem, "{data_atual}", agora.Format("02/01/2006"))
	
	return mensagem
}

// ===== PROCESSAMENTO EM BACKGROUND =====

// ProcessarAgendamentos processa agendamentos pendentes (deve ser chamado periodicamente)
func (s *RespostaRapidaService) ProcessarAgendamentos() error {
	agendamentos, err := s.repo.GetAgendamentosAtivos()
	if err != nil {
		return fmt.Errorf("erro ao buscar agendamentos ativos: %w", err)
	}

	for _, agendamento := range agendamentos {
		// Criar execução para o agendamento
		execucao := &models.ExecucaoResposta{
			RespostaRapidaID: agendamento.RespostaRapidaID,
			UsuarioID:        agendamento.UsuarioID,
			ChatID:           agendamento.ChatID,
			ContatoNome:      agendamento.ContatoNome,
			ContatoTelefone:  agendamento.ContatoTelefone,
			TriggerTipo:      agendamento.TriggerTipo,
			Status:           models.StatusPendente,
			TotalAcoes:       len(agendamento.RespostaRapida.Acoes),
		}

		err = s.repo.CreateExecucao(execucao)
		if err != nil {
			log.Printf("Erro ao criar execução para agendamento %s: %v", agendamento.ID, err)
			continue
		}

		// Executar
		go s.processarExecucao(execucao.ID)

		// Atualizar agendamento
		agendamento.ExecucoesRealizadas++
		
		if agendamento.ExecucoesRealizadas >= agendamento.MaxExecucoes {
			agendamento.Ativo = false
		} else if agendamento.RespostaRapida.IntervaloRepeticao != nil {
			proximaExecucao := time.Now().Add(time.Duration(*agendamento.RespostaRapida.IntervaloRepeticao) * time.Minute)
			agendamento.ProximaExecucao = &proximaExecucao
		}

		err = s.repo.UpdateAgendamento(&agendamento)
		if err != nil {
			log.Printf("Erro ao atualizar agendamento %s: %v", agendamento.ID, err)
		}
	}

	return nil
}

// ProcessarExecucoesPendentes processa execuções pendentes (deve ser chamado periodicamente)
func (s *RespostaRapidaService) ProcessarExecucoesPendentes() error {
	execucoes, err := s.repo.GetExecucoesPendentes()
	if err != nil {
		return fmt.Errorf("erro ao buscar execuções pendentes: %w", err)
	}

	for _, execucao := range execucoes {
		go s.processarExecucao(execucao.ID)
	}

	return nil
}

// ===== TIPOS DE REQUEST =====

type CreateRespostaRapidaRequest struct {
	Titulo                    string                     `json:"titulo"`
	Descricao                 *string                    `json:"descricao,omitempty"`
	CategoriaID               uuid.UUID                  `json:"categoria_id"`
	UsuarioID                 uuid.UUID                  `json:"usuario_id"`
	AgendamentoAtivo          bool                       `json:"agendamento_ativo"`
	TriggerTipo               models.TriggerTipo         `json:"trigger_tipo"`
	TriggerCondicao           *models.TriggerCondicao    `json:"trigger_condicao,omitempty"`
	DelaySegundos             int                        `json:"delay_segundos"`
	Repetir                   bool                       `json:"repetir"`
	IntervaloRepeticao        *int                       `json:"intervalo_repeticao,omitempty"`
	MaxRepeticoes             int                        `json:"max_repeticoes"`
	AplicarNovosContatos      bool                       `json:"aplicar_novos_contatos"`
	AplicarContatosExistentes bool                       `json:"aplicar_contatos_existentes"`
	ContatosEspecificos       []string                   `json:"contatos_especificos,omitempty"`
	Acoes                     []CreateAcaoRequest        `json:"acoes,omitempty"`
}

type CreateAcaoRequest struct {
	Tipo          models.TipoAcao          `json:"tipo"`
	DelaySegundos int                      `json:"delay_segundos"`
	Conteudo      models.ConteudoAcao      `json:"conteudo"`
	Obrigatorio   bool                     `json:"obrigatorio"`
	Condicional   bool                     `json:"condicional"`
}
