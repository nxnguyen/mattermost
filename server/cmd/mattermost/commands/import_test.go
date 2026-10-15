// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost/server/public/model"
)

func TestBulkImportApplyFlag(t *testing.T) {
	if testing.Short() {
		t.Skip("requires live test database")
	}

	// The subprocess exercises the real CLI/store. It only needs a config file,
	// not the additional mock App (and Redis) created by SetupWithStoreMock.
	dir := t.TempDir()
	th := &testHelper{tempDir: dir, configFilePath: filepath.Join(dir, "config.json")}
	config := &model.Config{}
	config.SetDefaults()
	*config.PluginSettings.Enable = false
	*config.AnnouncementSettings.AdminNoticesEnabled = false
	*config.AnnouncementSettings.UserNoticesEnabled = false
	th.SetConfig(config)

	// CLI commands use a read-only config store. Initialize app migrations on
	// this fresh test database first, as they may update configuration.
	a, err := initDBCommandContext(th.ConfigPath(), false)
	require.NoError(t, err)
	a.DoAppMigrations()
	a.Srv().Shutdown()

	db, err := sql.Open(model.DatabaseDriverPostgres, *mainHelper.GetSQLSettings().DataSource)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	for _, tc := range []struct {
		name       string
		flags      []string
		wantTeams  int
		wantOutput string
	}{
		{name: "default validates", wantOutput: "Validation complete."},
		{name: "validate does not persist", flags: []string{"--validate"}, wantOutput: "Validation complete."},
		{name: "apply persists", flags: []string{"--apply"}, wantTeams: 1, wantOutput: "Finished Bulk Import."},
		{name: "conflicting flags do not persist", flags: []string{"--apply", "--validate"}, wantOutput: "Use only one of --apply or --validate."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			teamName := model.NewRandomTeamName()
			file := filepath.Join(th.TemporaryDirectory(), teamName+".jsonl")
			data := fmt.Sprintf("{\"type\":\"version\",\"version\":1}\n{\"type\":\"team\",\"team\":{\"name\":%q,\"display_name\":\"Bulk CLI regression\",\"type\":\"O\"}}\n", teamName)
			require.NoError(t, os.WriteFile(file, []byte(data), 0600))

			args := append([]string{"import", "bulk", file}, tc.flags...)
			output, err := th.RunCommandWithOutput(t, args...)
			require.NoError(t, err, output)
			require.Contains(t, output, tc.wantOutput)

			var count int
			require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM Teams WHERE Name = $1", teamName).Scan(&count))
			require.Equal(t, tc.wantTeams, count, "--apply must persist, while validation must leave the database unchanged")
		})
	}
}
