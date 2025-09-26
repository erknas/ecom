package parse

import "strconv"

func UserID(idString string) (int64, error) {
	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		return 0, err
	}

	return id, nil
}
