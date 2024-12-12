package permission

type UserJob struct {
	UserId int64 `json:"userId,string" gorm:"comment:用户ID;not null;"` // 用户ID
	JobId  int64 `json:"JobId,string" gorm:"comment:岗位ID;not null;"`  // 岗位ID
}

func (UserJob) TableName() string {
	return "sys_user_job"
}
