package log

// Format 日志输出格式
type Format int

func (f Format) String() string {
	switch f {
	case TextFormat:
		return "text"
	case JsonFormat:
		return "json"
	}
	return "none"
}

const (
	TextFormat Format = iota // 文本格式
	JsonFormat               // JSON格式
)

// CutRule 日志切割规则
type CutRule int

const (
	CutByYear   CutRule = iota + 1 // 按照年切割
	CutByMonth                     // 按照月切割
	CutByDay                       // 按照日切割
	CutByHour                      // 按照时切割
	CutByMinute                    // 按照分切割
	CutBySecond                    // 按照秒切割
)

func (c CutRule) String() string {
	switch c {
	case CutByYear:
		return "year"
	case CutByMonth:
		return "month"
	case CutByDay:
		return "day"
	case CutByHour:
		return "hour"
	case CutByMinute:
		return "minute"
	case CutBySecond:
		return "second"
	}
	return "none"
}
