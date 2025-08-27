package services

import (
	"fmt"
	"log"
	"time"
	"tappyone/internal/models"

	"gorm.io/gorm"
)

// FluxoExecutionService gerencia a execução de fluxos de automação
type FluxoExecutionService struct {
	DB              *gorm.DB
	WhatsAppService *WhatsAppService
	KanbanService   *KanbanService
}

// ExecutionContext carrega o contexto de execução do fluxo
type ExecutionContext struct {
	FluxoID     string                 `json:"fluxo_id"`
	UserID      string                 `json:"user_id"`
	ContatoID   *string                `json:"contato_id,omitempty"`
	ChatID      *string                `json:"chat_id,omitempty"`
	CardID      *string                `json:"card_id,omitempty"`
	Variables   map[string]interface{} `json:"variables"`
	CurrentNode *models.FluxoNo        `json:"current_node"`
}

// NodeExecutionResult resultado da execução de um nó
type NodeExecutionResult struct {
	Success    bool                   `json:"success"`
	Error      *string                `json:"error,omitempty"`
	NextNodeID *string                `json:"next_node_id,omitempty"`
	Variables  map[string]interface{} `json:"variables"`
	Delay      *time.Duration         `json:"delay,omitempty"`
}

// NewFluxoExecutionService cria uma nova instância do serviço
func NewFluxoExecutionService(db *gorm.DB, whatsAppService *WhatsAppService, kanbanService *KanbanService) *FluxoExecutionService {
	return &FluxoExecutionService{
		DB:              db,
		WhatsAppService: whatsAppService,
		KanbanService:   kanbanService,
	}
}

// ExecuteFluxo executa um fluxo completo
func (s *FluxoExecutionService) ExecuteFluxo(fluxoID string, userID string, triggerData map[string]interface{}) error {
	log.Printf("[FLUXO] Executando fluxo %s para usuário %s", fluxoID, userID)

	// Carregar fluxo e nós
	var fluxo models.Fluxo
	if err := s.DB.Preload("Nos").Preload("Conexoes").Where("id = ? AND quadro_id IN (SELECT id FROM quadros WHERE usuario_id = ?)", fluxoID, userID).First(&fluxo).Error; err != nil {
		return fmt.Errorf("fluxo não encontrado: %w", err)
	}

	if !fluxo.Ativo {
		return fmt.Errorf("fluxo %s está inativo", fluxoID)
	}

	// Encontrar nó inicial (trigger)
	var startNode *models.FluxoNo
	for _, node := range fluxo.Nos {
		if node.Tipo == "trigger" {
			startNode = &node
			break
		}
	}

	if startNode == nil {
		return fmt.Errorf("nenhum nó de gatilho encontrado no fluxo %s", fluxoID)
	}

	// Criar contexto de execução
	context := &ExecutionContext{
		FluxoID:     fluxoID,
		UserID:      userID,
		Variables:   mergeMaps(triggerData, make(map[string]interface{})),
		CurrentNode: startNode,
	}

	// Extrair IDs do contexto de trigger
	if contatoID, ok := triggerData["contato_id"].(string); ok {
		context.ContatoID = &contatoID
	}
	if chatID, ok := triggerData["chat_id"].(string); ok {
		context.ChatID = &chatID
	}
	if cardID, ok := triggerData["card_id"].(string); ok {
		context.CardID = &cardID
	}

	// Executar fluxo
	return s.executeNode(context)
}

// executeNode executa um nó específico e continua o fluxo
func (s *FluxoExecutionService) executeNode(context *ExecutionContext) error {
	if context.CurrentNode == nil {
		log.Printf("[FLUXO] Execução concluída para fluxo %s", context.FluxoID)
		return nil
	}

	log.Printf("[FLUXO] Executando nó %s do tipo %s", context.CurrentNode.ID, context.CurrentNode.Tipo)

	// Executar nó baseado no tipo
	result, err := s.executeNodeByType(context)
	if err != nil {
		log.Printf("[FLUXO] Erro ao executar nó %s: %v", context.CurrentNode.ID, err)
		return err
	}

	if !result.Success {
		errorMsg := "erro desconhecido"
		if result.Error != nil {
			errorMsg = *result.Error
		}
		return fmt.Errorf("falha na execução do nó %s: %s", context.CurrentNode.ID, errorMsg)
	}

	// Atualizar variáveis do contexto
	context.Variables = mergeMaps(context.Variables, result.Variables)

	// Aplicar delay se especificado
	if result.Delay != nil {
		log.Printf("[FLUXO] Aguardando %v antes do próximo nó", *result.Delay)
		time.Sleep(*result.Delay)
	}

	// Continuar para o próximo nó
	if result.NextNodeID != nil {
		nextNode, err := s.findNodeByID(context.FluxoID, *result.NextNodeID)
		if err != nil {
			return fmt.Errorf("próximo nó não encontrado %s: %w", *result.NextNodeID, err)
		}
		context.CurrentNode = nextNode
		return s.executeNode(context)
	}

	log.Printf("[FLUXO] Nenhum próximo nó definido, execução concluída")
	return nil
}

// executeNodeByType executa um nó baseado no seu tipo
func (s *FluxoExecutionService) executeNodeByType(context *ExecutionContext) (*NodeExecutionResult, error) {
	switch context.CurrentNode.Tipo {
	case "trigger":
		return s.executeTriggerNode(context)
	case "condition":
		return s.executeConditionNode(context)
	case "action-chat":
		return s.executeChatActionNode(context)
	case "action-kanban":
		return s.executeKanbanActionNode(context)
	case "action-ia":
		return s.executeIAActionNode(context)
	case "action-resposta":
		return s.executeRespostaActionNode(context)
	case "action-agendamento":
		return s.executeAgendamentoActionNode(context)
	case "action-contrato":
		return s.executeContratoActionNode(context)
	case "action-delay":
		return s.executeDelayActionNode(context)
	case "action-webhook":
		return s.executeWebhookActionNode(context)
	case "action-database":
		return s.executeDatabaseActionNode(context)
	default:
		return nil, fmt.Errorf("tipo de nó não suportado: %s", context.CurrentNode.Tipo)
	}
}

// executeTriggerNode executa um nó de gatilho
func (s *FluxoExecutionService) executeTriggerNode(context *ExecutionContext) (*NodeExecutionResult, error) {
	// Gatilho simplesmente passa para o próximo nó
	nextNodeID, err := s.findNextNodeID(context.FluxoID, context.CurrentNode.ID)
	if err != nil {
		return nil, err
	}

	return &NodeExecutionResult{
		Success:    true,
		NextNodeID: nextNodeID,
		Variables:  make(map[string]interface{}),
	}, nil
}

// executeConditionNode executa um nó de condição
func (s *FluxoExecutionService) executeConditionNode(context *ExecutionContext) (*NodeExecutionResult, error) {
	config := map[string]interface{}(context.CurrentNode.Configuracao)

	// Avaliar condição (implementação simplificada)
	conditionResult := s.evaluateCondition(config, context.Variables)
	
	// Encontrar próximo nó baseado no resultado da condição
	var nextNodeID *string
	if conditionResult {
		if trueNode, exists := config["true_node_id"].(string); exists {
			nextNodeID = &trueNode
		}
	} else {
		if falseNode, exists := config["false_node_id"].(string); exists {
			nextNodeID = &falseNode
		}
	}

	return &NodeExecutionResult{
		Success:    true,
		NextNodeID: nextNodeID,
		Variables:  map[string]interface{}{"condition_result": conditionResult},
	}, nil
}

// executeChatActionNode executa ação de envio de mensagem
func (s *FluxoExecutionService) executeChatActionNode(context *ExecutionContext) (*NodeExecutionResult, error) {
	if context.ChatID == nil {
		return &NodeExecutionResult{
			Success: false,
			Error:   stringPtr("Chat ID não encontrado no contexto"),
		}, nil
	}

	config := map[string]interface{}(context.CurrentNode.Configuracao)

	message, exists := config["message"].(string)
	if !exists {
		return &NodeExecutionResult{
			Success: false,
			Error:   stringPtr("Mensagem não configurada"),
		}, nil
	}

	// Substituir variáveis na mensagem
	processedMessage := s.replaceVariables(message, context.Variables)

	// Enviar mensagem via WhatsApp
	sessionName := fmt.Sprintf("user_%s", context.UserID)
	_, err := s.WhatsAppService.SendMessage(sessionName, *context.ChatID, processedMessage)
	if err != nil {
		return &NodeExecutionResult{
			Success: false,
			Error:   stringPtr(fmt.Sprintf("Erro ao enviar mensagem: %v", err)),
		}, nil
	}

	nextNodeID, _ := s.findNextNodeID(context.FluxoID, context.CurrentNode.ID)
	return &NodeExecutionResult{
		Success:    true,
		NextNodeID: nextNodeID,
		Variables:  map[string]interface{}{"message_sent": processedMessage},
	}, nil
}

// executeKanbanActionNode executa ação no Kanban
func (s *FluxoExecutionService) executeKanbanActionNode(context *ExecutionContext) (*NodeExecutionResult, error) {
	if context.CardID == nil {
		return &NodeExecutionResult{
			Success: false,
			Error:   stringPtr("Card ID não encontrado no contexto"),
		}, nil
	}

	config := map[string]interface{}(context.CurrentNode.Configuracao)

	targetColumnID, exists := config["target_column_id"].(string)
	if !exists {
		return &NodeExecutionResult{
			Success: false,
			Error:   stringPtr("Coluna de destino não configurada"),
		}, nil
	}

	// Mover card (implementação simplificada)
	err := s.moveCard(*context.CardID, targetColumnID, context.UserID)
	if err != nil {
		return &NodeExecutionResult{
			Success: false,
			Error:   stringPtr(fmt.Sprintf("Erro ao mover card: %v", err)),
		}, nil
	}

	nextNodeID, _ := s.findNextNodeID(context.FluxoID, context.CurrentNode.ID)
	return &NodeExecutionResult{
		Success:    true,
		NextNodeID: nextNodeID,
		Variables:  map[string]interface{}{"card_moved": true, "new_column_id": targetColumnID},
	}, nil
}

// executeDelayActionNode executa ação de delay
func (s *FluxoExecutionService) executeDelayActionNode(context *ExecutionContext) (*NodeExecutionResult, error) {
	config := map[string]interface{}(context.CurrentNode.Configuracao)

	delaySeconds, exists := config["delay_seconds"].(float64)
	if !exists {
		delaySeconds = 5 // Default 5 seconds
	}

	delay := time.Duration(delaySeconds) * time.Second
	nextNodeID, _ := s.findNextNodeID(context.FluxoID, context.CurrentNode.ID)

	return &NodeExecutionResult{
		Success:    true,
		NextNodeID: nextNodeID,
		Variables:  map[string]interface{}{"delay_applied": delaySeconds},
		Delay:      &delay,
	}, nil
}

// Implementações placeholder para outros tipos de nós
func (s *FluxoExecutionService) executeIAActionNode(context *ExecutionContext) (*NodeExecutionResult, error) {
	// TODO: Implementar processamento com IA
	nextNodeID, _ := s.findNextNodeID(context.FluxoID, context.CurrentNode.ID)
	return &NodeExecutionResult{
		Success:    true,
		NextNodeID: nextNodeID,
		Variables:  map[string]interface{}{"ia_processed": true},
	}, nil
}

func (s *FluxoExecutionService) executeRespostaActionNode(context *ExecutionContext) (*NodeExecutionResult, error) {
	// TODO: Implementar execução de resposta rápida
	nextNodeID, _ := s.findNextNodeID(context.FluxoID, context.CurrentNode.ID)
	return &NodeExecutionResult{
		Success:    true,
		NextNodeID: nextNodeID,
		Variables:  map[string]interface{}{"resposta_executed": true},
	}, nil
}

func (s *FluxoExecutionService) executeAgendamentoActionNode(context *ExecutionContext) (*NodeExecutionResult, error) {
	// TODO: Implementar criação de agendamento
	nextNodeID, _ := s.findNextNodeID(context.FluxoID, context.CurrentNode.ID)
	return &NodeExecutionResult{
		Success:    true,
		NextNodeID: nextNodeID,
		Variables:  map[string]interface{}{"agendamento_created": true},
	}, nil
}

func (s *FluxoExecutionService) executeContratoActionNode(context *ExecutionContext) (*NodeExecutionResult, error) {
	// TODO: Implementar geração de contrato
	nextNodeID, _ := s.findNextNodeID(context.FluxoID, context.CurrentNode.ID)
	return &NodeExecutionResult{
		Success:    true,
		NextNodeID: nextNodeID,
		Variables:  map[string]interface{}{"contrato_generated": true},
	}, nil
}

func (s *FluxoExecutionService) executeWebhookActionNode(context *ExecutionContext) (*NodeExecutionResult, error) {
	// TODO: Implementar chamada de webhook
	nextNodeID, _ := s.findNextNodeID(context.FluxoID, context.CurrentNode.ID)
	return &NodeExecutionResult{
		Success:    true,
		NextNodeID: nextNodeID,
		Variables:  map[string]interface{}{"webhook_called": true},
	}, nil
}

func (s *FluxoExecutionService) executeDatabaseActionNode(context *ExecutionContext) (*NodeExecutionResult, error) {
	// TODO: Implementar operações de banco de dados
	nextNodeID, _ := s.findNextNodeID(context.FluxoID, context.CurrentNode.ID)
	return &NodeExecutionResult{
		Success:    true,
		NextNodeID: nextNodeID,
		Variables:  map[string]interface{}{"database_updated": true},
	}, nil
}

// Funções auxiliares

func (s *FluxoExecutionService) findNodeByID(fluxoID, nodeID string) (*models.FluxoNo, error) {
	var node models.FluxoNo
	if err := s.DB.Where("id = ? AND fluxo_id = ?", nodeID, fluxoID).First(&node).Error; err != nil {
		return nil, err
	}
	return &node, nil
}

func (s *FluxoExecutionService) findNextNodeID(fluxoID, currentNodeID string) (*string, error) {
	var conexao models.FluxoConexao
	if err := s.DB.Where("de_id = ?", currentNodeID).First(&conexao).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Não há próximo nó
		}
		return nil, err
	}
	return &conexao.ParaID, nil
}

func (s *FluxoExecutionService) evaluateCondition(config map[string]interface{}, variables map[string]interface{}) bool {
	// Implementação simplificada de avaliação de condições
	// TODO: Implementar parser de condições mais robusto
	
	field, hasField := config["field"].(string)
	operator, hasOperator := config["operator"].(string)
	value, hasValue := config["value"]

	if !hasField || !hasOperator || !hasValue {
		return false
	}

	variableValue, exists := variables[field]
	if !exists {
		return false
	}

	switch operator {
	case "equals":
		return variableValue == value
	case "not_equals":
		return variableValue != value
	case "contains":
		if str, ok := variableValue.(string); ok {
			if searchStr, ok := value.(string); ok {
				return contains(str, searchStr)
			}
		}
	case "greater_than":
		if num1, ok := variableValue.(float64); ok {
			if num2, ok := value.(float64); ok {
				return num1 > num2
			}
		}
	case "less_than":
		if num1, ok := variableValue.(float64); ok {
			if num2, ok := value.(float64); ok {
				return num1 < num2
			}
		}
	}

	return false
}

func (s *FluxoExecutionService) replaceVariables(message string, variables map[string]interface{}) string {
	result := message
	for key, value := range variables {
		placeholder := fmt.Sprintf("{%s}", key)
		if str, ok := value.(string); ok {
			result = replaceAll(result, placeholder, str)
		} else {
			result = replaceAll(result, placeholder, fmt.Sprintf("%v", value))
		}
	}
	return result
}

func (s *FluxoExecutionService) moveCard(cardID, columnID, userID string) error {
	// Implementação simplificada de movimentação de card
	// TODO: Usar KanbanService quando disponível
	return s.DB.Model(&models.Card{}).Where("id = ? AND quadro_id IN (SELECT id FROM quadros WHERE usuario_id = ?)", cardID, userID).Update("coluna_id", columnID).Error
}

// Funções utilitárias
func mergeMaps(map1, map2 map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range map1 {
		result[k] = v
	}
	for k, v := range map2 {
		result[k] = v
	}
	return result
}

func stringPtr(s string) *string {
	return &s
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || s[0:len(substr)] == substr || contains(s[1:], substr))
}

func replaceAll(s, old, new string) string {
	// Implementação simples de replace
	for i := 0; i <= len(s)-len(old); i++ {
		if s[i:i+len(old)] == old {
			return s[:i] + new + replaceAll(s[i+len(old):], old, new)
		}
	}
	return s
}
