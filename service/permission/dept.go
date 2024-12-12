package permission

import (
	"errors"
	"github.com/tealeg/xlsx"
	"go-apevolo/global"
	"go-apevolo/model"
	"go-apevolo/model/permission"
	"go-apevolo/payloads/dto"
	"go-apevolo/payloads/request"
	"go-apevolo/utils"
	"go-apevolo/utils/ext"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"strconv"
)

type DeptService struct{}

// Create
// @description: 创建
// @receiver: deptService
// @param: req
// @return: error
func (deptService *DeptService) Create(req *dto.CreateUpdateDeptDto) error {
	dept := &permission.Department{}
	var total int64
	err := global.Db.Model(&permission.Department{}).Scopes(utils.IsDeleteSoft).Where("name = ?", req.Name).Count(&total).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	if total > 0 {
		return errors.New("部门名称=>" + req.Name + "=>已存在!")
	}
	db := global.Db.Begin()
	defer func() {
		if err != nil {
			global.Logger.Error("tran error: ", zap.Error(err))
			db.Rollback()
		} else {
			db.Commit()
		}
	}()
	req.Generate(dept)
	err = db.Create(dept).Error
	if err == nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	if req.ParentId != 0 {
		//重新计算子部门个数
		parentDept := &permission.Department{}
		err = db.Scopes(utils.IsDeleteSoft).Where("id = ?", req.ParentId).First(parentDept).Error
		if err != nil {
			return err
		}
		err = db.Model(&permission.Department{}).Scopes(utils.IsDeleteSoft).Where("parent_id = ?", parentDept.Id).Count(&total).Error
		if err != nil {
			global.Logger.Error("db error: ", zap.Error(err))
			return err
		}
		parentDept.SubCount = int(total)
		err = db.Updates(parentDept).Error
		if err != nil {
			global.Logger.Error("db error: ", zap.Error(err))
			return err
		}
	}
	return nil
}

// Update
// @description: 更新
// @receiver: deptService
// @param: req
// @return: error
func (deptService *DeptService) Update(req *dto.CreateUpdateDeptDto) error {
	var total int64
	oldDept := &permission.Department{}
	dept := &permission.Department{}
	err := global.Db.Scopes(utils.IsDeleteSoft).Where("id = ?", req.Id).First(oldDept).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("数据不存在或您无权查看！")
		}
		return err
	}
	if oldDept.Name != req.Name {
		err = global.Db.Model(&permission.Department{}).Scopes(utils.IsDeleteSoft).Where("name = ?", req.Name).Count(&total).Error
		if err != nil {
			global.Logger.Error("db error: ", zap.Error(err))
			return err
		}
		if total > 0 {
			return errors.New("部门名称=>" + req.Name + "=>已存在!")
		}
	}
	db := global.Db.Begin()
	defer func() {
		if err != nil {
			global.Logger.Error("tran error: ", zap.Error(err))
			db.Rollback()
		} else {
			db.Commit()
		}
	}()
	req.Generate(dept)
	dept.SubCount = oldDept.SubCount
	err = db.Save(dept).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	if oldDept.ParentId != req.ParentId {
		if req.ParentId != 0 {
			dept := &permission.Department{}
			err = global.Db.Scopes(utils.IsDeleteSoft).Where("id = ?", req.ParentId).First(dept).Error
			if err != nil {
				global.Logger.Error("db error: ", zap.Error(err))
				return err
			}
			err = global.Db.Model(&permission.Department{}).Where("parent_id = ? ", dept.Id).Count(&total).Error
			if err != nil {
				global.Logger.Error("db error: ", zap.Error(err))
				return err
			}
			dept.SubCount = int(total)
			err = global.Db.Updates(dept).Error
			if err != nil {
				global.Logger.Error("db error: ", zap.Error(err))
				return err
			}
		}

		if oldDept.ParentId != 0 {
			dept := &permission.Department{}
			err = global.Db.Scopes(utils.IsDeleteSoft).Where("id = ?", oldDept.ParentId).First(dept).Error
			if err != nil {
				global.Logger.Error("db error: ", zap.Error(err))
				return err
			}
			err = global.Db.Model(&permission.Department{}).Where("parent_id = ? ", dept.Id).Count(&total).Error
			if err != nil {
				global.Logger.Error("db error: ", zap.Error(err))
				return err
			}
			dept.SubCount = int(total)
			err = global.Db.Updates(dept).Error
			if err != nil {
				global.Logger.Error("db error: ", zap.Error(err))
				return err
			}
		}
	}
	return nil
}

// Delete
// @description: 删除
// @receiver: deptService
// @param: ids
// @param: updateBy
// @return: error
func (deptService *DeptService) Delete(ids []int64, updateBy string) error {
	allIds, err := deptService.GetChildrenIds(ids, []int64{})
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}

	departmentList := make([]permission.Department, 0)
	err = deptService.getDepartmentList(&allIds, &departmentList)
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	if len(departmentList) == 0 {
		return errors.New("数据不存在")
	}
	for i := range departmentList {
		if len(departmentList[i].Users) > 0 {
			return errors.New("存在用户关联，请解除后再试！")
		}
		if len(departmentList[i].Roles) > 0 {
			return errors.New("存在角色关联，请解除后再试！")
		}
	}

	localTime := ext.GetCurrentTime()
	err = global.Db.Model(&permission.Department{}).Scopes(utils.IsDeleteSoft).Where("id in (?)", allIds).Updates(
		permission.Department{BaseModel: model.BaseModel{
			UpdateBy:   &updateBy,
			UpdateTime: &localTime,
		}, SoftDeleted: model.SoftDeleted{IsDeleted: true}},
	).Error

	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}

	var pIds []int64
	for _, dept := range departmentList {
		pIds = utils.AppendInt64(pIds, dept.ParentId)
	}
	updateDeptList := make([]permission.Department, 0)
	err = global.Db.Model(&permission.Department{}).Scopes(utils.IsDeleteSoft).Where("id in (?)", pIds).Find(updateDeptList).Error
	if err != nil {
		return err
	}
	if len(updateDeptList) > 0 {
		for _, dept := range updateDeptList {
			var total int64
			err = global.Db.Model(&permission.Department{}).Scopes(utils.IsDeleteSoft).Where("parent_id = ? ", dept.Id).Count(&total).Error
			if err != nil {
				global.Logger.Error("db error: ", zap.Error(err))
				return err
			}
			dept.SubCount = int(total)
			err = global.Db.Model(&permission.Department{}).Where("id = ?", dept.Id).UpdateColumn("sub_count", dept.SubCount).Error
			if err != nil {
				global.Logger.Error("db error: ", zap.Error(err))
				return err
			}
		}
		//err = global.Db.Model(&permission.Department{}).Updates(updateDeptList).Error
	}
	return nil
}

// GetChildrenIds
// @description: 获取所有子集ID
// @receiver: deptService
// @param: ids
// @param: allIds
// @return: allIdList
// @return: err
func (deptService *DeptService) GetChildrenIds(ids []int64, allIds []int64) (allIdList []int64, err error) {
	for _, id := range ids {
		allIds = utils.AppendInt64(allIds, id)
		var depts []permission.Department
		err = global.Db.Scopes(utils.IsDeleteSoft).Where("enabled = 1 and parent_id = ? ", id).Find(&depts).Error
		if err == nil && len(depts) > 0 {
			var ids []int64
			for _, d := range depts {
				ids = append(ids, d.Id)
			}
			allIds, err = deptService.GetChildrenIds(ids, allIds)
			if err != nil {
				//return allIds, err
				break
			}
		}
	}
	return allIds, err
}

// Superior
// @description: 查询同级与父级部门
// @receiver: deptService
// @param: id
// @return: list
// @return: total
// @return: err
func (deptService *DeptService) Superior(id int64) (list interface{}, total int64, err error) {
	// 创建db
	//db := global.Db.Model(&permission.Department{})
	var deptMap = make(map[int64][]permission.Department)

	var dept permission.Department
	err = global.Db.Scopes(utils.IsDeleteSoft).Where("id = ?", id).First(&dept).Error
	if err != nil {
		return
	}
	deptMap, err = deptService.superior(dept, []permission.Department{})
	var depts = deptMap[0]
	for i := 0; i < len(depts); i++ {
		err = deptService.getChildrenList(&depts[i], deptMap)
	}
	//if err == nil {
	//	for i := range depts {
	//		depts[i].CalculateHasChildren()
	//		depts[i].CalculateLeaf()
	//		depts[i].CalculateLabel()
	//	}
	//}
	total = int64(len(depts))
	return depts, total, err
}

func (deptService *DeptService) superior(dept permission.Department, departments []permission.Department) (deptMap map[int64][]permission.Department, err error) {
	var deptTmpList []permission.Department
	deptMap = make(map[int64][]permission.Department)
	if dept.ParentId == 0 {
		err = global.Db.Scopes(utils.IsDeleteSoft).Where("parent_id = 0  ").Find(&deptTmpList).Error
		if err != nil {
			return
		}
		departments = append(departments, deptTmpList...)
		for _, v := range departments {
			deptMap[v.ParentId] = append(deptMap[v.ParentId], v)
		}
	} else {
		err = global.Db.Scopes(utils.IsDeleteSoft).Where("parent_id = ? ", dept.ParentId).Find(&deptTmpList).Error
		if err != nil {
			return
		}
		departments = append(departments, deptTmpList...)
		var deptTmp permission.Department
		err = global.Db.Scopes(utils.IsDeleteSoft).Where("id = ? ", dept.ParentId).First(&deptTmp).Error
		if err != nil {
			return
		}
		deptMap, err = deptService.superior(deptTmp, departments)
	}
	return deptMap, err
}

func (deptService *DeptService) getChildrenList(dept *permission.Department, treeMap map[int64][]permission.Department) (err error) {
	dept.CalculateHasChildren()
	dept.CalculateLeaf()
	dept.CalculateLabel()
	dept.Children = treeMap[dept.Id]
	for i := 0; i < len(dept.Children); i++ {
		err = deptService.getChildrenList(&dept.Children[i], treeMap)
	}
	return err
}

// Query
// @description: 查询
// @receiver: deptService
// @param: info
// @param: list
// @param: count
// @return: error
func (deptService *DeptService) Query(info *dto.DeptQueryCriteria, list *[]permission.Department, count *int64) error {
	if info.Pagination.PageSize == 0 {
		pagination := request.NewPagination()
		info.Pagination = *pagination
	}
	limit := info.PageSize
	offset := info.PageSize * (info.PageIndex - 1)
	// 创建db
	db := buildDeptQuery(global.Db.Model(&permission.Department{}), info)

	var err error
	if info.ParentId == nil {
		err = db.Count(count).Limit(limit).Offset(offset).Find(list).Error
	} else {
		err = db.Count(count).Find(list).Error
	}
	//err := db.Count(total).Limit(limit).Offset(offset).Find(list).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}

// Download
// @description: 导出
// @receiver: deptService
// @param: info
// @return: filePath
// @return: fileName
// @return: err
func (deptService *DeptService) Download(info *dto.DeptQueryCriteria) (filePath string, fileName string, err error) {
	var depts []permission.Department
	// 创建db并构建查询条件
	err = buildDeptQuery(global.Db.Model(&permission.Department{}), info).Find(&depts).Error

	if err != nil {
		return
	}
	// 创建一个新的 Excel 文件
	file := xlsx.NewFile()
	sheet, err := file.AddSheet("Depts")
	if err != nil {
		return
	}
	style := xlsx.NewStyle()
	style.Fill = *xlsx.NewFill("solid", "1e90ff", "1e90ff")
	style.ApplyFill = true
	// 表头数据
	header := []string{"ID", "部门名称", "部门父ID", "排序", "是否启用", "子部门个数", "创建时间"}
	// 将表头添加到 Excel 工作表
	row := sheet.AddRow()
	for _, cellValue := range header {
		cell := row.AddCell()
		cell.Value = cellValue
		cell.SetStyle(style)
	}

	// 将数据添加到 Excel 工作表
	for _, dept := range depts {
		row := sheet.AddRow()
		row.WriteSlice(&[]interface{}{
			strconv.FormatInt(dept.Id, 10),
			dept.Name,
			dept.ParentId,
			dept.Sort,
			func() string {
				if dept.Enabled {
					return "启用"
				}
				return "禁用"
			}(),
			dept.SubCount,
			dept.CreateTime.Format("2006-01-02 15:04:05"),
		}, -1)
	}
	// 保存 Excel 文件到本地
	fileName = "Depts_" + utils.GenerateID().String() + ".xlsx"
	filePath = global.Config.Excel.Dir + fileName
	err = file.Save(filePath)
	return filePath, fileName, err
}

// getDepartmentList
// @description: 获取部门列表
// @receiver: deptService
// @param: ids
// @param: departmentList
// @return: error
func (deptService *DeptService) getDepartmentList(ids *[]int64, departmentList *[]permission.Department) error {
	err := global.Db.Where("enabled = 1 and id in (?)", *ids).Scopes(utils.IsDeleteSoft).Preload("Roles").Preload("Users").Find(departmentList).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}

// buildDeptQuery
// @description: 条件表达式
// @param: db
// @param: info
// @return: *gorm.DB
func buildDeptQuery(db *gorm.DB, info *dto.DeptQueryCriteria) *gorm.DB {
	if info.ParentId != nil {
		db = db.Where("parent_id = ? ", info.ParentId)
	} else {
		db = db.Where("parent_id = 0 ")
	}
	if info.DeptName != "" {
		db = db.Where("name LIKE ?", "%"+info.DeptName+"%")
	}
	if info.Enabled != nil {
		db = db.Where("enabled = ?", info.Enabled)
	}
	if len(info.CreateTime) > 0 {
		db = db.Where("create_time >= ? and create_time <= ? ", info.CreateTime[0], info.CreateTime[1])
	}
	if len(info.SortFields) > 0 {
		for _, sort := range info.SortFields {
			db = db.Order(sort)
		}
	}
	db = db.Scopes(utils.IsDeleteSoft)

	return db
}
