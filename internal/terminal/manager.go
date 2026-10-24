package terminal

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"
)

var (
	adjectives = []string{
		"admiring", "affectionate", "agitated", "amazing", "angry", "awesome",
		"blissful", "bold", "brave", "busy", "charming", "clever", "cool",
		"eager", "ecstatic", "elastic", "elated", "elegant", "eloquent",
		"epic", "exciting", "fervent", "festive", "flamboyant", "focused",
		"friendly", "frosty", "funny", "gallant", "gifted", "goofy", "gracious",
		"great", "happy", "heuristic", "hopeful", "hungry", "infallible",
		"inspiring", "jolly", "jovial", "keen", "kind", "laughing", "loving",
		"lucid", "magical", "mystifying", "modest", "musing", "naughty",
		"nervous", "nice", "objective", "optimistic", "peaceful", "pedantic",
		"pensive", "practical", "priceless", "proud", "quizzical", "recursing",
		"relaxed", "reverent", "ridiculous", "robotic", "romantic", "sad",
		"sage", "serene", "sharp", "silly", "sleepy", "stoic", "stupefied",
		"suspicious", "sweet", "tender", "thirsty", "trusting", "unruffled",
		"upbeat", "vibrant", "vigilant", "vigorous", "wizardly", "wonderful",
		"xenodochial", "youthful", "zealous",
	}
	nouns = []string{
		"albattani", "allen", "almeida", "antonelli", "archimedes",
		"ardinghelli", "aryabhata", "austin", "babbage", "banach", "banesa",
		"bardeen", "bartik", "bassi", "beaver", "bell", "benz", "bhabha",
		"bhaskara", "blackburn", "blackwell", "bohr", "booth", "borg", "bose",
		"boyd", "brahmagupta", "brattain", "brown", "buck", "burnell", "cajal",
		"cannon", "carson", "cartwright", "chandrasekhar", "chaplygin", "chatterjee",
		"clarke", "colden", "cori", "cray", "curie", "darwin", "davinci",
		"deborah", "diffie", "dirac", "driscoll", "dubinsky", "easley", "edison",
		"einstein", "elbakyan", "elgamal", "elion", "ellis", "engelbart",
		"euler", "faraday", "feistel", "fermat", "fermi", "feynman", "franklin",
		"gagarin", "galileo", "galois", "gates", "gauss", "germain", "goldberg",
		"goldstine", "goldwasser", "golick", "goodall", "gould", "grasshopper",
		"greider", "grothendieck", "haibt", "hamilton", "haslett", "hawking",
		"heisenberg", "hermann", "herschel", "hertz", "heyrovsky", "higgs",
		"hipparchus", "hirohito", "hopper", "hugle", "hypatia", "ishizaka",
		"jackson", "jang", "jepsen", "johnson", "joliot", "jones", "kalam",
		"kare", "keldysh", "keller", "kepler", "khayyam", "kilby", "kirch",
		"knuth", "kowalevski", "lalande", "lamarr", "lamport", "leakey",
		"leavitt", "lewin", "lichterman", "liskov", "lovelace",
		"lumiere", "mahavira", "montalcini", "moser", "napier", "nash", "neumann",
		"newton", "nightingale", "nobel", "noether", "northcutt", "noyce",
		"panini", "pare", "pascal", "pasteur", "payne", "perlman", "pike",
		"poincare", "poitras", "proskuriakova", "ptolemy", "raman", "ramanujan",
		"ride", "ritchie", "roentgen", "rosalind", "rubin", "saha", "sammet",
		"sanderson", "satoshi", "shamir", "shannon", "shirley", "shockley",
		"shtern", "sinan", "snyder", "solomon", "spence", "stallman", "stonebraker",
		"sutherland", "swartz", "swirles", "taussig", "terry", "tesla", "tharp",
		"thompson", "torvalds", "turing", "varahamihira", "vaughan", "villani",
		"visvesvaraya", "volhard", "wescoff", "wilbur", "wiles", "williams",
		"williamson", "wilson", "wing", "wozniak", "wright", "yalow",
		"yong", "zhuse",
	}
)

type TerminalManager struct {
	mu        sync.Mutex
	terminals map[string]*ManagedTerminal
}

func NewTerminalManager() *TerminalManager {
	return &TerminalManager{
		terminals: make(map[string]*ManagedTerminal),
	}
}

func (m *TerminalManager) Create(cfg CreateConfig) (*ManagedTerminal, error) {
	id := m.uniqueID()

	mt, err := NewManagedTerminal(id, cfg)
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if existing, ok := m.terminals[id]; ok {
		existing.Close()
	}
	m.terminals[id] = mt
	return mt, nil
}

func (m *TerminalManager) uniqueID() string {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := 0; i < 1000; i++ {
		id := generateRandomName()
		if _, taken := m.terminals[id]; !taken {
			return id
		}
	}
	return fmt.Sprintf("terminal_%d", time.Now().UnixNano())
}

func (m *TerminalManager) Get(id string) (*ManagedTerminal, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	mt, ok := m.terminals[id]
	if !ok {
		return nil, fmt.Errorf("terminal %q not found", id)
	}
	return mt, nil
}

type TerminalInfo struct {
	ID        string `json:"id"`
	Cols      uint16 `json:"cols"`
	Rows      uint16 `json:"rows"`
	Command   string `json:"command"`
	Recording bool   `json:"recording"`
}

func (m *TerminalManager) List() []TerminalInfo {
	m.mu.Lock()
	defer m.mu.Unlock()

	infos := make([]TerminalInfo, 0, len(m.terminals))
	for _, mt := range m.terminals {
		mt.mu.Lock()
		infos = append(infos, TerminalInfo{
			ID:        mt.ID,
			Cols:      mt.Cols,
			Rows:      mt.Rows,
			Command:   mt.Command,
			Recording: mt.Recorder != nil,
		})
		mt.mu.Unlock()
	}
	return infos
}

func (m *TerminalManager) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	mt, ok := m.terminals[id]
	if !ok {
		return fmt.Errorf("terminal %q not found", id)
	}

	mt.Close()
	delete(m.terminals, id)
	return nil
}

func (m *TerminalManager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, mt := range m.terminals {
		mt.Close()
		delete(m.terminals, id)
	}
}

func generateRandomName() string {
	adjIndex, _ := rand.Int(rand.Reader, big.NewInt(int64(len(adjectives))))
	nounIndex, _ := rand.Int(rand.Reader, big.NewInt(int64(len(nouns))))
	return adjectives[adjIndex.Int64()] + "_" + nouns[nounIndex.Int64()]
}
