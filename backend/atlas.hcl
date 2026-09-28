data "external_schema" "gorm" {
  program = [
    "go",
    "run",
    "-mod=mod",
    "ariga.io/atlas-provider-gorm",
    "load",
    "--path", "./internal/models",
    "--dialect", "postgres",
  ]
}

env "gorm" {
  src = data.external_schema.gorm.url
  dev = "docker://postgres/17/dev?search_path=public"

  migration {
    dir    = "file://internal/migration/migrations"
    format = "golang-migrate"
  }

  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}
