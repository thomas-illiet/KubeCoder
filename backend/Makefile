.PHONY: atlas-schema atlas-diff atlas-hash atlas-validate

ATLAS ?= atlas

atlas-schema:
	$(ATLAS) schema inspect --env gorm --url env://src

atlas-diff:
	@test -n "$(NAME)" || (echo "NAME is required, for example: make atlas-diff NAME=add_organizations" && exit 1)
	$(ATLAS) migrate diff --env gorm "$(NAME)"

atlas-hash:
	$(ATLAS) migrate hash --env gorm

atlas-validate:
	$(ATLAS) migrate validate --env gorm
