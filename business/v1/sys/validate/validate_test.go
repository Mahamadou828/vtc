package validate_test

import (
	"testing"

	"github.com/google/uuid"
	"vtc/business/v1/sys/validate"
)

func Test_Check(t *testing.T) {
	t.Logf("Given the need to validate struct type")
	{
		s := struct {
			Name  string `validate:"required"`
			Email string `validate:"required,email"`
		}{
			Name:  "Mahamadou",
			Email: "test@test.com",
		}
		t.Logf("\tTest %v:\tWhen handling a valid struct", s)
		{

			if err := validate.Check(&s); err != nil {
				t.Fatalf("\tTest: %s:\tShould validate the struct type: %s", s, err)
			}
		}

		w := struct {
			Name  string `validate:"required"`
			Email string `validate:"required,email"`
		}{}
		t.Logf("\tTest %v:\tWhen handling an invalid struct", w)
		{

			err := validate.Check(&w)
			if err == nil {
				t.Fatalf("\tTest: %s:\tShould return an error for invalid struct: %s", s, err)
			}
		}
	}
}

func Test_GenerateID(t *testing.T) {
	t.Logf("Given the need to generate uuid")
	{
		s := struct {
			ID string `validate:"required,uuid"`
		}{
			ID: validate.GenerateID(),
		}

		t.Logf("\tTest %v:\tWhen handling a valid uuid", s)
		{
			if err := validate.Check(&s); err != nil {
				t.Fatalf("\tTest: %s:\tShould validate the uuid: %s", s, err)
			}
		}

		w := struct {
			ID string `validate:"required,uuid"`
		}{
			ID: "TEST",
		}
		t.Logf("\tTest %v:\tWhen handling an invalid uuid", w)
		{
			err := validate.Check(&w)
			if err == nil {
				t.Fatalf("\tTest: %s:\tShould return an error for invalid uuid: %s", s, err)
			}
		}
	}
}

func Test_CheckID(t *testing.T) {
	t.Logf("Given the need to check id")
	{
		id := uuid.NewString()
		t.Logf("\tTest %v:\tcheck generated id", id)
		{
			if err := validate.CheckID(id); err != nil {
				t.Fatalf("\tTest %v:\tgiven uuid should be valid %v", id, err)
			}
		}
	}
}

func Test_CheckEmail(t *testing.T) {
	t.Logf("Given the need to check email")
	{
		vEmail := "test@gmail.com"
		invalidEmail := "invalid"
		t.Logf("\tTest %v:\tgiven email should be valid", vEmail)
		{
			if !validate.CheckEmail(vEmail) {
				t.Logf("\tTest %v:\tgiven email should be valid", vEmail)
			}
		}
		t.Logf("\tTest %v:\tgiven email should be invalid", vEmail)
		{
			if validate.CheckEmail(invalidEmail) {
				t.Logf("\tTest %v:\tgiven email should be invalid", vEmail)
			}
		}
	}
}
