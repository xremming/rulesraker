package archiver

import (
	"time"
)

type JSONDate time.Time

func (jd JSONDate) String() string {
	return time.Time(jd).Format("2006-01-02")
}

func (jd *JSONDate) UnmarshalJSON(data []byte) error {
	t, err := time.Parse(`"2006-01-02"`, string(data))
	if err != nil {
		return err
	}

	*jd = JSONDate(t)
	return nil
}

func (jd JSONDate) MarshalJSON() ([]byte, error) {
	return []byte(time.Time(jd).Format(`"2006-01-02"`)), nil
}
