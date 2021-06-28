package util


 
/**
获取本周周一的日期
*/
func GetFirstDateOfWeek() (weekMonday string) {
	now := time.Now()
 
	offset := int(time.Monday - now.Weekday())
	if offset > 0 {
		offset = -6
	}
 
	weekStartDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, offset)
	weekMonday = weekStartDate.Format("2006-01-02")
	return
}
 
/**
获取上周的周一日期
*/
func GetLastWeekFirstDate() (weekMonday string) {
	thisWeekMonday := GetFirstDateOfWeek()
	TimeMonday, _ := time.Parse("2006-01-02", thisWeekMonday)
	lastWeekMonday := TimeMonday.AddDate(0, 0, -7)
	weekMonday = lastWeekMonday.Format("2006-01-02")
	return
}
————————————————
版权声明：本文为CSDN博主「wangxiaoangg」的原创文章，遵循CC 4.0 BY-SA版权协议，转载请附上原文出处链接及本声明。
原文链接：https://blog.csdn.net/qq_16399991/article/details/98599756