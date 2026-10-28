package server

var stacks = [30]string{
	"python",
	"C",
	"C++",
	"java",
	"C#",
	"javascript",
	"docker",
	"r",
	"sql",
	"rust",
	"go",
	"php",
	"html",
	"css",
	"scratch",
	"bash",
	"ruby",
	"ia",
	"typescript",
	"nodejs",
	"react",
	"angular",
	"postgresql",
	"kubernetes",
	"mysql",
	"sqlite",
	"git",
	"django",
	"laravel",
	"nextjs",
}

func checkCategories(categories []string) bool {

	founds := make([]bool, len(categories))

	for i := 0; i < len(categories); i++ {
		for j := 0; j < len(stacks); j++ {
			if categories[i] == stacks[j] {
				founds[i] = true
				continue
			}
		}
	}

	for i := 0; i < len(founds); i++ {
		if founds[i] == false {
			return false
		}
	}

	return true
}
