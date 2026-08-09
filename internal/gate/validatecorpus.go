package gate

import (
	"fmt"
	"strings"

	"github.com/iderex/messbuch/internal/schema"
	"github.com/iderex/messbuch/internal/validate"
)

// validateCorpusLeg refuses a file under record/ that is not a well formed
// record.
//
// It answers the structural question only: is this a record at all. The
// authority is schema/record-<n>.toml rather than this source, and the leg
// prints how many records it read against how many versions, so a run over an
// empty corpus cannot be read as a run that checked something.
//
// It fails closed in the two directions that matter. A schema set it cannot
// read is a refusal rather than a corpus validated against nothing, and a
// record directory it cannot walk is a refusal rather than an empty corpus.
func validateCorpusLeg(root string) (string, error) {
	set, err := schema.Load(root)
	if err != nil {
		return "", err
	}
	report, err := validate.Corpus(root, set)
	if err != nil {
		return "", err
	}

	if len(report.Refusals) > 0 {
		var lines []string
		for _, r := range report.Refusals {
			lines = append(lines, r.String())
		}
		return "", fmt.Errorf("%d refusal(s) over %d record(s):\n  %s\n\nEvery refusal this leg can produce is printed by:\n\n    go run . refusals",
			len(report.Refusals), report.Records, strings.Join(lines, "\n  "))
	}

	examined := fmt.Sprintf("%d record(s) under %s/ against schema version(s) %s",
		report.Records, validate.RecordDir, versions(set))
	if report.Skipped > 0 {
		examined += fmt.Sprintf("; %d file(s) under %s were not read",
			report.Skipped, strings.Join(report.Excluded, ", "))
	}
	return examined, nil
}

// versions renders the schema versions the leg read, so the line says what the
// corpus was measured against rather than that it was measured.
func versions(set *schema.Set) string {
	var out []string
	for _, v := range set.SortedVersions() {
		out = append(out, fmt.Sprint(v))
	}
	return strings.Join(out, ", ")
}

// limitsOfValidateCorpus is printed beside the leg's result, because an
// assurance whose edge is one document away gets quoted without it.
const limitsOfValidateCorpus = `this is the structural question only: whether a file under record/ is a well
formed record against the schema it names. It reads one file at a time and
never opens a second to decide the first, so a quantity with no vocabulary
entry, a group identifier resolving to no registry file, a supersession
naming a record that does not exist and a conversion factor that is the wrong
number all pass here. Those are the meaning leg, which is not built and is
owed by #25. A refusal names the field's path inside the file rather than a
line number, except a parse failure, which carries the line the decoder
reported.`

// Sites is the refusal catalogue, so that the command that prints it does not
// import the validator's internals a second time.
func Sites() []validate.Site { return validate.Catalogue }
