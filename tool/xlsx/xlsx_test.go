package xlsx

import "testing"

func TestBuildData(t *testing.T) {
	fileName := "demo.xlsx"
	header := []string{"姓名", "年龄", "性别"}
	list := [][]string{
		[]string{"ChamPly", "19", "M"},
		[]string{"Tom", "18", "W"},
		[]string{"CC", "", "S"},
		[]string{"C", "S", ""},
	}

	err := WriteData(list, header, fileName)
	if err != nil {
		t.Errorf("构建demo.xlsx失败:err:%+v", err)
	}

}

func TestReadData(t *testing.T) {
	data, err := ReadData("demo.xlsx")
	if err != nil {
		t.Errorf("读取demo.xlsx失败:err:%+v", err)
	}

	dataInput := [][]string{
		[]string{"姓名", "年龄", "性别"},
		[]string{"ChamPly", "19", "M"},
		[]string{"Tom", "18", "W"},
		[]string{"CC", "", "S"},
		[]string{"C", "S", ""},
	}

	for i, in1 := range data {
		for j, in2 := range dataInput {
			if i != j {
				continue
			}
			if !compare(in1, in2) {
				t.Errorf("第 %d 行数据不一致, in1:%+v, in2:%+v", i, in1, in2)
			}
			break
		}
	}
}

func compare(in1 []string, in2 []string) bool {
	if len(in1) != len(in2) {
		return false
	}

	if len(in1) == 0 {
		return true
	}

	for i := 0; i < len(in1); i++ {
		if in1[i] != in2[i] {
			return false
		}
	}
	return true
}
