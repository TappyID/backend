package repositories

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"tappyone/internal/models"
)

type RespostaRapidaRepository struct {
	db *gorm.DB
}

func NewRespostaRapidaRepository(db *gorm.DB) *RespostaRapidaRepository {
	return &RespostaRapidaRepository{db: db}
}

// ===== CATEGORIAS =====

// CreateCategoria cria uma nova categoria
func (r *RespostaRapidaRepository) CreateCategoria(categoria *models.CategoriaResposta) error {
	return r.db.Create(categoria).Error
}

// GetCategoriasByUsuario busca categorias por usuário
func (r *RespostaRapidaRepository) GetCategoriasByUsuario(usuarioID uuid.UUID) ([]models.CategoriaResposta, error) {
	var categorias []models.CategoriaResposta
	err := r.db.Where("usuario_id = ? AND ativo = ?", usuarioID, true).
		Order("ordem ASC, nome ASC").
		Preload("RespostasRapidas", "ativo = true").
		Find(&categorias).Error
	return categorias, err
}

// GetCategoriaByID busca categoria por ID
func (r *RespostaRapidaRepository) GetCategoriaByID(id uuid.UUID) (*models.CategoriaResposta, error) {
	var categoria models.CategoriaResposta
	err := r.db.Where("id = ?", id).
		Preload("RespostasRapidas", "ativo = true").
		First(&categoria).Error
	if err != nil {
		return nil, err
	}
	return &categoria, nil
}

// UpdateCategoria atualiza uma categoria
func (r *RespostaRapidaRepository) UpdateCategoria(categoria *models.CategoriaResposta) error {
	return r.db.Save(categoria).Error
}

// DeleteCategoria exclui uma categoria (soft delete)
func (r *RespostaRapidaRepository) DeleteCategoria(id uuid.UUID) error {
	return r.db.Model(&models.CategoriaResposta{}).
		Where("id = ?", id).
		Update("ativo", false).Error
}

// ===== RESPOSTAS RÁPIDAS =====

// CreateRespostaRapida cria uma nova resposta rápida
func (r *RespostaRapidaRepository) CreateRespostaRapida(resposta *models.RespostaRapida) error {
	return r.db.Create(resposta).Error
}

// GetRespostasRapidasByUsuario busca respostas rápidas por usuário
func (r *RespostaRapidaRepository) GetRespostasRapidasByUsuario(usuarioID uuid.UUID) ([]models.RespostaRapida, error) {
	var respostas []models.RespostaRapida
	err := r.db.Where("usuario_id = ? AND ativo = ?", usuarioID, true).
		Order("ordem ASC, titulo ASC").
		Preload("Categoria").
		Preload("Acoes", "ativo = true", func(db *gorm.DB) *gorm.DB {
			return db.Order("ordem ASC")
		}).
		Find(&respostas).Error
	return respostas, err
}

// GetRespostasRapidasByCategoria busca respostas rápidas por categoria
func (r *RespostaRapidaRepository) GetRespostasRapidasByCategoria(categoriaID uuid.UUID) ([]models.RespostaRapida, error) {
	var respostas []models.RespostaRapida
	err := r.db.Where("categoria_id = ? AND ativo = ?", categoriaID, true).
		Order("ordem ASC, titulo ASC").
		Preload("Categoria").
		Preload("Acoes", "ativo = true", func(db *gorm.DB) *gorm.DB {
			return db.Order("ordem ASC")
		}).
		Find(&respostas).Error
	return respostas, err
}

// GetRespostaRapidaByID busca resposta rápida por ID
func (r *RespostaRapidaRepository) GetRespostaRapidaByID(id uuid.UUID) (*models.RespostaRapida, error) {
	var resposta models.RespostaRapida
	err := r.db.Where("id = ?", id).
		Preload("Categoria").
		Preload("Acoes", "ativo = true", func(db *gorm.DB) *gorm.DB {
			return db.Order("ordem ASC")
		}).
		First(&resposta).Error
	if err != nil {
		return nil, err
	}
	return &resposta, nil
}

// GetRespostasRapidasAtivas busca respostas rápidas ativas para agendamento
func (r *RespostaRapidaRepository) GetRespostasRapidasAtivas(usuarioID uuid.UUID) ([]models.RespostaRapida, error) {
	var respostas []models.RespostaRapida
	err := r.db.Where("usuario_id = ? AND ativo = ? AND pausado = ? AND agendamento_ativo = ?", 
		usuarioID, true, false, true).
		Preload("Categoria").
		Preload("Acoes", "ativo = true", func(db *gorm.DB) *gorm.DB {
			return db.Order("ordem ASC")
		}).
		Find(&respostas).Error
	return respostas, err
}

// UpdateRespostaRapida atualiza uma resposta rápida
func (r *RespostaRapidaRepository) UpdateRespostaRapida(resposta *models.RespostaRapida) error {
	return r.db.Save(resposta).Error
}

// DeleteRespostaRapida exclui uma resposta rápida (soft delete)
func (r *RespostaRapidaRepository) DeleteRespostaRapida(id uuid.UUID) error {
	return r.db.Model(&models.RespostaRapida{}).
		Where("id = ?", id).
		Update("ativo", false).Error
}

// TogglePausarRespostaRapida pausa/despausa uma resposta rápida
func (r *RespostaRapidaRepository) TogglePausarRespostaRapida(id uuid.UUID, pausado bool) error {
	return r.db.Model(&models.RespostaRapida{}).
		Where("id = ?", id).
		Update("pausado", pausado).Error
}

// ===== AÇÕES =====

// CreateAcao cria uma nova ação
func (r *RespostaRapidaRepository) CreateAcao(acao *models.AcaoResposta) error {
	return r.db.Create(acao).Error
}

// GetAcoesByResposta busca ações por resposta rápida
func (r *RespostaRapidaRepository) GetAcoesByResposta(respostaID uuid.UUID) ([]models.AcaoResposta, error) {
	var acoes []models.AcaoResposta
	err := r.db.Where("resposta_rapida_id = ? AND ativo = ?", respostaID, true).
		Order("ordem ASC").
		Find(&acoes).Error
	return acoes, err
}

// UpdateAcao atualiza uma ação
func (r *RespostaRapidaRepository) UpdateAcao(acao *models.AcaoResposta) error {
	return r.db.Save(acao).Error
}

// DeleteAcao exclui uma ação (soft delete)
func (r *RespostaRapidaRepository) DeleteAcao(id uuid.UUID) error {
	return r.db.Model(&models.AcaoResposta{}).
		Where("id = ?", id).
		Update("ativo", false).Error
}

// ReorderAcoes reordena as ações de uma resposta rápida
func (r *RespostaRapidaRepository) ReorderAcoes(respostaID uuid.UUID, acoesOrdem []uuid.UUID) error {
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for i, acaoID := range acoesOrdem {
		if err := tx.Model(&models.AcaoResposta{}).
			Where("id = ? AND resposta_rapida_id = ?", acaoID, respostaID).
			Update("ordem", i).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// ===== EXECUÇÕES =====

// CreateExecucao cria uma nova execução
func (r *RespostaRapidaRepository) CreateExecucao(execucao *models.ExecucaoResposta) error {
	return r.db.Create(execucao).Error
}

// GetExecucoesByUsuario busca execuções por usuário
func (r *RespostaRapidaRepository) GetExecucoesByUsuario(usuarioID uuid.UUID, limit int) ([]models.ExecucaoResposta, error) {
	var execucoes []models.ExecucaoResposta
	query := r.db.Where("usuario_id = ?", usuarioID).
		Order("created_at DESC").
		Preload("RespostaRapida").
		Preload("RespostaRapida.Categoria")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	err := query.Find(&execucoes).Error
	return execucoes, err
}

// GetExecucoesPendentes busca execuções pendentes para processamento
func (r *RespostaRapidaRepository) GetExecucoesPendentes() ([]models.ExecucaoResposta, error) {
	var execucoes []models.ExecucaoResposta
	err := r.db.Where("status = ? AND (agendado_para IS NULL OR agendado_para <= ?)", 
		models.StatusPendente, time.Now()).
		Order("created_at ASC").
		Preload("RespostaRapida").
		Preload("RespostaRapida.Acoes", "ativo = true", func(db *gorm.DB) *gorm.DB {
			return db.Order("ordem ASC")
		}).
		Find(&execucoes).Error
	return execucoes, err
}

// UpdateExecucao atualiza uma execução
func (r *RespostaRapidaRepository) UpdateExecucao(execucao *models.ExecucaoResposta) error {
	return r.db.Save(execucao).Error
}

// ===== AGENDAMENTOS =====

// CreateAgendamento cria um novo agendamento
func (r *RespostaRapidaRepository) CreateAgendamento(agendamento *models.AgendamentoResposta) error {
	return r.db.Create(agendamento).Error
}

// GetAgendamentosByUsuario busca agendamentos por usuário
func (r *RespostaRapidaRepository) GetAgendamentosByUsuario(usuarioID uuid.UUID) ([]models.AgendamentoResposta, error) {
	var agendamentos []models.AgendamentoResposta
	err := r.db.Where("usuario_id = ? AND ativo = ?", usuarioID, true).
		Order("created_at DESC").
		Preload("RespostaRapida").
		Preload("RespostaRapida.Categoria").
		Find(&agendamentos).Error
	return agendamentos, err
}

// GetAgendamentosAtivos busca agendamentos ativos para processamento
func (r *RespostaRapidaRepository) GetAgendamentosAtivos() ([]models.AgendamentoResposta, error) {
	var agendamentos []models.AgendamentoResposta
	err := r.db.Where("ativo = ? AND pausado = ? AND (proxima_execucao IS NULL OR proxima_execucao <= ?)", 
		true, false, time.Now()).
		Preload("RespostaRapida").
		Preload("RespostaRapida.Acoes", "ativo = true", func(db *gorm.DB) *gorm.DB {
			return db.Order("ordem ASC")
		}).
		Find(&agendamentos).Error
	return agendamentos, err
}

// GetAgendamentoByChat busca agendamento por chat e resposta
func (r *RespostaRapidaRepository) GetAgendamentoByChat(respostaID uuid.UUID, chatID string, usuarioID uuid.UUID) (*models.AgendamentoResposta, error) {
	var agendamento models.AgendamentoResposta
	err := r.db.Where("resposta_rapida_id = ? AND chat_id = ? AND usuario_id = ?", 
		respostaID, chatID, usuarioID).
		First(&agendamento).Error
	if err != nil {
		return nil, err
	}
	return &agendamento, nil
}

// UpdateAgendamento atualiza um agendamento
func (r *RespostaRapidaRepository) UpdateAgendamento(agendamento *models.AgendamentoResposta) error {
	return r.db.Save(agendamento).Error
}

// DeleteAgendamento exclui um agendamento
func (r *RespostaRapidaRepository) DeleteAgendamento(id uuid.UUID) error {
	return r.db.Delete(&models.AgendamentoResposta{}, id).Error
}

// TogglePausarAgendamento pausa/despausa um agendamento
func (r *RespostaRapidaRepository) TogglePausarAgendamento(id uuid.UUID, pausado bool) error {
	return r.db.Model(&models.AgendamentoResposta{}).
		Where("id = ?", id).
		Update("pausado", pausado).Error
}

// ===== ESTATÍSTICAS =====

// GetEstatisticasUsuario busca estatísticas de um usuário
func (r *RespostaRapidaRepository) GetEstatisticasUsuario(usuarioID uuid.UUID) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Total de categorias
	var totalCategorias int64
	r.db.Model(&models.CategoriaResposta{}).
		Where("usuario_id = ? AND ativo = ?", usuarioID, true).
		Count(&totalCategorias)
	stats["total_categorias"] = totalCategorias
	
	// Total de respostas rápidas
	var totalRespostas int64
	r.db.Model(&models.RespostaRapida{}).
		Where("usuario_id = ? AND ativo = ?", usuarioID, true).
		Count(&totalRespostas)
	stats["total_respostas"] = totalRespostas
	
	// Total de respostas ativas
	var respostasAtivas int64
	r.db.Model(&models.RespostaRapida{}).
		Where("usuario_id = ? AND ativo = ? AND pausado = ? AND agendamento_ativo = ?", 
			usuarioID, true, false, true).
		Count(&respostasAtivas)
	stats["respostas_ativas"] = respostasAtivas
	
	// Total de execuções
	var totalExecucoes int64
	r.db.Model(&models.ExecucaoResposta{}).
		Where("usuario_id = ?", usuarioID).
		Count(&totalExecucoes)
	stats["total_execucoes"] = totalExecucoes
	
	// Execuções hoje
	hoje := time.Now().Truncate(24 * time.Hour)
	var execucoesHoje int64
	r.db.Model(&models.ExecucaoResposta{}).
		Where("usuario_id = ? AND created_at >= ?", usuarioID, hoje).
		Count(&execucoesHoje)
	stats["execucoes_hoje"] = execucoesHoje
	
	return stats, nil
}
