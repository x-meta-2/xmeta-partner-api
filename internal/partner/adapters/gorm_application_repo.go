package adapters

import (
	"xmeta-partner/controllers/common"
	"xmeta-partner/database"
	"xmeta-partner/structs"

	"gorm.io/gorm"
)

type GormApplicationRepo struct {
	DB *gorm.DB
}

func (r *GormApplicationRepo) FindByID(id string) (*database.PartnerApplication, error) {
	var app database.PartnerApplication
	if err := r.DB.Preload("User").Preload("Reviewer").Where("id = ?", id).First(&app).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *GormApplicationRepo) FindPendingByID(id string) (*database.PartnerApplication, error) {
	var app database.PartnerApplication
	if err := r.DB.Where("id = ? AND status = ?", id, database.ApplicationStatusPending).First(&app).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *GormApplicationRepo) List(params structs.ApplicationListParams) ([]database.PartnerApplication, int, error) {
	pInput := common.PreparePaginationInput(params.PaginationInput)
	params.PaginationInput = pInput

	orm := r.DB.Model(&database.PartnerApplication{}).Preload("User").Preload("Reviewer")
	orm = common.Equal(orm, "status", params.Status)

	if params.Query != "" {
		q := "%" + params.Query + "%"
		orm = orm.
			Joins("LEFT JOIN users ON users.id = partner_applications.user_id").
			Where(
				"users.email ILIKE ? OR users.first_name ILIKE ? OR users.last_name ILIKE ? OR partner_applications.company_name ILIKE ?",
				q, q, q, q,
			)
	}

	total := common.Total(orm.Scopes(common.SortDateFilter(&params.PaginationInput)))

	var apps []database.PartnerApplication
	if err := orm.
		Order("created_at desc").
		Scopes(common.Paginate(&params.PaginationInput)).
		Find(&apps).Error; err != nil {
		return nil, 0, err
	}

	return apps, total, nil
}

func (r *GormApplicationRepo) Save(app *database.PartnerApplication) error {
	return r.DB.Save(app).Error
}
