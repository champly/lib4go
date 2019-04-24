package xlsx

import (
	"fmt"

	"github.com/champly/lib4go/file"
	"github.com/tealeg/xlsx"
)

func WriteData(list [][]string, header []string, fileName string) (err error) {
	file := xlsx.NewFile()
	sheet, err := file.AddSheet("Sheet1")
	if err != nil {
		return fmt.Errorf("add sheet fail:%s", err)
	}

	row := sheet.AddRow()

	// add header
	addData(header, row)

	// add content
	for _, ctx := range list {
		row = sheet.AddRow()
		addData(ctx, row)
	}

	if err = file.Save(fileName); err != nil {
		return fmt.Errorf("save xlsx file:%s fail:%s", fileName, err)
	}
	return nil
}

func addData(data []string, row *xlsx.Row) {
	var cell *xlsx.Cell
	for _, col := range data {
		cell = row.AddCell()
		cell.Value = col
	}
}

func ReadData(fileName string) (list [][]string, err error) {
	if !file.Exists(fileName) {
		return nil, fmt.Errorf("filename:%s is not exists", fileName)
	}

	xlf, err := xlsx.OpenFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("open xlsx:%s err:%s", fileName, err)
	}

	for _, sheet := range xlf.Sheets {
		for i, row := range sheet.Rows {
			data := []string{}
			var l int
			for j, cell := range row.Cells {
				// record header lenght
				if i == 0 {
					l++
				}
				if j > l-1 {
					break
				}

				str := cell.String()
				data = append(data, str)
			}

			if len(data) < 1 {
				break
			}
			list = append(list, data)
		}
	}
	return list, nil
}
