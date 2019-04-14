package killgrave

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

func imposterHandler(imposter Imposter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := validateSchema(imposter, r.Body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		w.Header().Set("Content-Type", imposter.Response.ContentType)
		w.WriteHeader(imposter.Response.Status)
		writeBody(imposter, w)
	}
}

func validateSchema(imposter Imposter, bodyRequest io.ReadCloser) error {
	if imposter.Request.SchemaFile == nil {
		return nil
	}

	schemaFile := *imposter.Request.SchemaFile
	if _, err := os.Stat(schemaFile); os.IsNotExist(err) {
		return errors.Wrapf(err, "the schema file %s not found", schemaFile)
	}

	b, err := ioutil.ReadAll(bodyRequest)
	if err != nil {
		return errors.Wrapf(err, "impossible read the request body")
	}

	dir, _ := os.Getwd()
	schemaFilePath := "file://" + dir + "/" + schemaFile
	schema := gojsonschema.NewReferenceLoader(schemaFilePath)
	document := gojsonschema.NewStringLoader(string(b))

	res, err := gojsonschema.Validate(schema, document)
	if err != nil {
		return errors.Wrap(err, "error validating the json schema")
	}

	if !res.Valid() {
		for _, desc := range res.Errors() {
			return errors.New(desc.String())
		}
	}

	return nil

}

func writeBody(imposter Imposter, w http.ResponseWriter) {
	wb := []byte(imposter.Response.Body)

	if imposter.Response.BodyFile != nil {
		wb = fetchBodyFromFile(*imposter.Response.BodyFile)
	}
	w.Write(wb)
}

func fetchBodyFromFile(bodyFile string) (bytes []byte) {
	if _, err := os.Stat(bodyFile); os.IsNotExist(err) {
		log.Printf("the body file %s not found\n", bodyFile)
		return
	}

	f, _ := os.Open(bodyFile)
	defer f.Close()
	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		log.Printf("imposible read the file %s: %v\n", bodyFile, err)
	}
	return
}
