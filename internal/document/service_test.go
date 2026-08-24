package document

import (
	"strings"
	"testing"

	"github.com/wavefnd/wave-platform/internal/storage"
)

func TestOfficialDocumentationUsesStableLanguageVocabulary(t *testing.T) {
	documents, _, err := readOfficialDocuments()
	if err != nil {
		t.Fatal(err)
	}

	forbidden := []string{
		"`let`",
		"`let mut`",
		"i256",
		"i512",
		"i1024",
		"u256",
		"u512",
		"u1024",
		"current compiler contract",
		"current implementation",
		"this release",
		"removed syntax",
		"code generation treats",
		"lexer recognizes",
		"type-conversion path",
		"현재 컴파일러 계약",
		"현재 구현",
		"이 릴리스",
		"폐지된 문법",
		"코드 생성 단계",
		"타입 변환 경로",
	}
	for _, document := range documents {
		for _, phrase := range forbidden {
			if strings.Contains(document.Markdown, phrase) {
				t.Errorf("%s/%s contains implementation-history wording %q", document.Locale, document.Path, phrase)
			}
		}
	}

	expected := map[string][]string{
		"en": {
			"`var` is the syntax for declaring local variables.",
			"`i8`, `i16`, `i32`, `i64`, `i128`",
			"`isz`",
			"`usz`",
			"**Wave Explicit Memory Type Model**",
		},
		"ko": {
			"`var`는 지역 변수를 선언하는 문법입니다.",
			"`i8`, `i16`, `i32`, `i64`, `i128`",
			"`isz`",
			"`usz`",
			"**Wave Explicit Memory Type Model**",
		},
	}
	for locale, phrases := range expected {
		var combined strings.Builder
		for _, document := range documents {
			if document.Locale == locale {
				combined.WriteString(document.Markdown)
			}
		}
		for _, phrase := range phrases {
			if !strings.Contains(combined.String(), phrase) {
				t.Errorf("%s documentation is missing %q", locale, phrase)
			}
		}
	}

	requiredByDocument := map[string][]string{
		"en/language/declarations-and-types": {
			"A type alias is a readable alternative name for another type.",
			"`isz` and `usz` represent signed and unsigned integers sized for the target's address space.",
		},
		"ko/language/declarations-and-types": {
			"타입 별칭은 같은 타입을 코드의 문맥에 맞는 이름으로 표현하는 문법입니다.",
			"`isz`는 주소 크기에 맞는 부호 있는 정수 타입이고, `usz`는 주소 크기에 맞는 부호 없는 정수 타입입니다.",
		},
		"en/language/explicit-memory-type-model": {
			"`null` represents a pointer that does not point to a value.",
			"pointers and arrays as explicit, language-level memory types",
		},
		"ko/language/explicit-memory-type-model": {
			"`null`은 유효한 메모리 주소를 가리키지 않는 포인터 값입니다.",
			"언어 차원의 명시적인 메모리 타입",
		},
		"en/language/console-io-and-formatting": {
			"Arrays and structs are not formatting arguments.",
		},
		"ko/language/console-io-and-formatting": {
			"배열과 구조체는 포매팅 인자로 사용할 수 없습니다.",
		},
		"en/language/modules-imports-and-ffi": {
			"Local modules use a path beginning with `./`.",
		},
		"ko/language/modules-imports-and-ffi": {
			"로컬 파일 경로는 `./`로 시작",
		},
	}
	for _, document := range documents {
		key := document.Locale + "/" + document.Path
		for _, phrase := range requiredByDocument[key] {
			if !strings.Contains(document.Markdown, phrase) {
				t.Errorf("%s is missing stable language contract %q", key, phrase)
			}
		}
	}
}

func TestSeedOfficialPublishesSupportedDocumentationLocales(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })

	count, err := SeedOfficial(database)
	if err != nil || count != 61 {
		t.Fatalf("seed count=%d err=%v", count, err)
	}
	repository := NewRepository(database)
	expectedTitles := map[string]string{
		"en": "Pointers and explicit memory access",
		"ko": "포인터와 명시적 메모리 접근",
	}
	for _, locale := range []string{"en", "ko"} {
		items, err := repository.Summaries(locale)
		if err != nil || len(items) != 27 {
			t.Fatalf("%s summaries=%d err=%v", locale, len(items), err)
		}
		view, err := repository.Published(locale, "language/explicit-memory-type-model")
		if err != nil {
			t.Fatal(err)
		}
		if view.Title != expectedTitles[locale] {
			t.Fatalf("unexpected memory document: %#v", view)
		}
		if !strings.Contains(view.Markdown, "ptr<T>") || !strings.Contains(view.Markdown, "```wave") {
			t.Fatal("published revision did not preserve the Markdown authoring source")
		}
	}
	for _, locale := range []string{"ja", "zh", "es", "de", "ru", "id", "vi"} {
		items, err := repository.Summaries(locale)
		if err != nil || len(items) != 1 || items[0].Path != "getting-started/overview" {
			t.Fatalf("%s summaries=%#v err=%v", locale, items, err)
		}
	}
	install, err := repository.Published("en", "getting-started/install")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(install.Markdown, "curl -fsSL https://wave-lang.dev/install.sh | bash -s -- latest") {
		t.Fatal("official installer command is missing")
	}
	count, err = SeedOfficial(database)
	if err != nil || count != 0 {
		t.Fatalf("second seed count=%d err=%v", count, err)
	}
}

func TestOfficialInstallDocumentsIncludeWindowsInstaller(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if _, err := SeedOfficial(database); err != nil {
		t.Fatal(err)
	}
	repository := NewRepository(database)
	for _, locale := range []string{"en", "ko"} {
		install, err := repository.Published(locale, "getting-started/install")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(install.Markdown, "https://wave-lang.dev/install.ps1") || !strings.Contains(install.Markdown, "-Latest") {
			t.Fatalf("%s official Windows installer command is missing", locale)
		}
		if !strings.Contains(install.Markdown, "vex --version") || !strings.Contains(install.Markdown, "VexVersion") {
			t.Fatalf("%s Vex installation guidance is missing", locale)
		}
	}
}
