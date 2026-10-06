// Package catalogtest builds a small synthetic data directory for tests. It contains no Nissan data.
package catalogtest

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite" // registers the "sqlite" database/sql driver
)

// Schema mirrors data/data.db after the optimisation pass.
const Schema = `
CREATE TABLE catalog (etd TEXT, grupo TEXT, model TEXT, cmodel TEXT, drive TEXT, date_from TEXT, date_to TEXT,
  serie TEXT, data TEXT, orden TEXT, langs TEXT);
CREATE TABLE vin (vin TEXT, vin_raw TEXT, codenis TEXT, etd TEXT, tipo TEXT, prodata TEXT, vin_rev TEXT);
CREATE TABLE grupo (etd TEXT, grupo TEXT, esp TEXT);
CREATE TABLE main_group (etd TEXT, cetd TEXT, lang TEXT, label TEXT);
CREATE TABLE section (etd TEXT, lang TEXT, plate TEXT, codigrup TEXT, numsec TEXT, nomsec TEXT, notas TEXT,
  desde TEXT, hasta TEXT);
CREATE TABLE part (etd TEXT, lang TEXT, pospie INTEGER, plate TEXT, item TEXT, item_eff TEXT, sec TEXT,
  nc TEXT, s TEXT, ind1 TEXT, ind2 TEXT, ind3 TEXT, ind4 TEXT, ind5 TEXT,
  part_no TEXT, part_key TEXT, des TEXT, esp TEXT, qty TEXT, cap TEXT, pmc TEXT, app TEXT,
  dataplic TEXT, ica TEXT, ree TEXT, ree_key TEXT, kdf TEXT, pnc TEXT);
CREATE TABLE attr_table (etd TEXT, grupo TEXT, kind TEXT, tabla TEXT, lang TEXT, label TEXT);
CREATE TABLE attr_value (etd TEXT, grupo TEXT, tabla TEXT, cetd TEXT, code TEXT, lang TEXT, label TEXT, orden TEXT);
CREATE TABLE modelnis (etd TEXT, codenis TEXT, grupo TEXT,
  c01 TEXT, c02 TEXT, c03 TEXT, c04 TEXT, c05 TEXT, c06 TEXT, c07 TEXT, c08 TEXT, c09 TEXT, c10 TEXT);
CREATE TABLE infosec (etd TEXT, variant TEXT, secc TEXT, grupo TEXT, codigrup TEXT, desdat TEXT, finsdat TEXT,
  c01 TEXT, c02 TEXT, c03 TEXT, c04 TEXT, c05 TEXT, c06 TEXT, c07 TEXT, c08 TEXT, c09 TEXT, c10 TEXT);
CREATE TABLE hotspot (etd TEXT, kind TEXT, image TEXT, caption TEXT, x INTEGER, y INTEGER, w INTEGER, h INTEGER);
CREATE TABLE cinfo (etd TEXT, file TEXT);
CREATE VIRTUAL TABLE part_fts USING fts5(des, part_no, part_key, pnc, content='part', content_rowid='rowid');
CREATE INDEX vin_vin ON vin(vin);
CREATE INDEX vin_rev ON vin(vin_rev);
CREATE INDEX part_key ON part(part_key);
CREATE INDEX part_ree_key ON part(ree_key);
`

const rows = `
INSERT INTO catalog VALUES
 ('AA','G01','VANETTE','C220L','LHD','04/87','11/94','0049','0049 TEST','N001','es,en,de,fr'),
 ('AB','G01','VANETTE','C220R','RHD','07/87','09/94','0050','0050 TEST','N002','en');
INSERT INTO modelnis VALUES
 -- c04..c10 are NULL like in the real data (infosec uses '-' there, modelnis never does).
 ('AA','BELC220QSKVX','G01','1','1','1',NULL,NULL,NULL,NULL,NULL,NULL,NULL),
 ('AA','BELC220QSKVX','G01','1','1','1',NULL,NULL,NULL,NULL,NULL,NULL,NULL), -- duplicate row, as in the real data
 ('AA','BELC220QJKL','G01','2','2','1',NULL,NULL,NULL,NULL,NULL,NULL,NULL),
 ('AB','RMODEL','G01','1',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL);
INSERT INTO attr_table VALUES
 ('AA','G01','T','T01','en','ENGINE'),('AA','G01','T','T01','fr','MOTEUR'),
 ('AA','G01','T','T02','en','WHEEL BASE'),('AA','G01','T','T02','fr','EMPATTEMENT'),
 ('AA','G01','I','I01','en','BODY'),('AA','G01','I','I01','fr',NULL),
 ('AB','G01','T','T01','en','ENGINE');
INSERT INTO attr_value VALUES
 ('AA','G01','T01','1','LD20','en','LD20',NULL),('AA','G01','T01','1','LD20','fr','LD20',NULL),
 ('AA','G01','T01','2','A15','en','A15',NULL),('AA','G01','T01','2','A15','fr','A15',NULL),
 ('AA','G01','T02','1','SHORT','en','SHORT',NULL),('AA','G01','T02','1','SHORT','fr','COURT',NULL),
 ('AA','G01','T02','2','LONG','en','LONG',NULL),('AA','G01','T02','2','LONG','fr','LONG',NULL),
 ('AA','G01','I01','1','VAN','en','VAN',NULL),('AA','G01','I01','1','VAN','fr',NULL,NULL),
 ('AB','G01','T01','1','LD20','en','LD20',NULL);
INSERT INTO main_group VALUES
 ('AA','A','en','ENGINE'),('AA','A','fr','MOTEUR'),
 ('AA','B','en','ENGINE ELECTRICAL SYSTEM'),('AA','B','fr','SYSTEME ELECTRIQUE MOTEUR'),
 ('AB','A','en','ENGINE');
INSERT INTO section VALUES
 ('AA','en','AA101','A','101','ENGINE ASSY','LD20','04-87','12-95'),
 ('AA','en','AA230','B','230','ALTERNATOR FITTING',NULL,'04-87','12-88'),
 ('AA','en','AA230A','B','230A','ALTERNATOR FITTING','LD20-II','04-87','12-95'),
 ('AA','en','AA231','B','231','ALTERNATOR',NULL,'04-87','12-95'),
 ('AA','fr','AA101','A','101','MOTEUR COMPLET','LD20','04-87','12-95'),
 ('AA','fr','AA230','B','230','FIXATION ALTERNATEUR',NULL,'04-87','12-88'),
 ('AA','fr','AA230A','B','230A','FIXATION ALTERNATEUR','LD20-II','04-87','12-95'),
 ('AA','fr','AA231','B','231','ALTERNATEUR',NULL,'04-87','12-95'),
 ('AB','en','AB040','A','040','ENGINE',NULL,'01-87','12-94');
INSERT INTO infosec VALUES
 ('AA','F','101','G01','A','198704','199512','0','0','0','-','-','-','-','-','-','-'),
 ('AA','F','230','G01','B','198704','198812','1','0','0','-','-','-','-','-','-','-'),
 ('AA','F','230A','G01','B','198704','199512','1','0','0','-','-','-','-','-','-','-'),
 ('AA','F','231','G01','B','198704','199512','2','0','0','-','-','-','-','-','-','-'),
 ('AA','','230','G01','B','198704','199512','0','0','0','-','-','-','-','-','-','-'),
 ('AB','F','040','G01','A','198701','199412','0','-','-','-','-','-','-','-','-','-');
INSERT INTO hotspot VALUES
 ('AA','plate','AA230A','1',100,100,20,20),
 ('AA','plate','AA230A','2',200,100,20,20),
 ('AA','plate','AA230A','2',300,150,20,20),
 ('AA','group','B','230',10,10,40,20),
 ('AA','group','B','231',60,10,40,20);
INSERT INTO cinfo VALUES ('AA','G0101.pdf');
`

type vinRow struct{ vin, raw, model, etd, prod string }

var vins = []vinRow{
	{"VSKBEC220U0990494", "VSKBEC220U0990494", "BELC220QSKVX", "AA", "198905"},
	{"VSKBEC220U0990494", "VSKBEC220U0990494", "BELC220QSKVX", "AA", "198905"},
	{"VSKBEC220U0111111", "VSKBEC220U0111111", "BELC220QJKL", "AA", "199301"},
	{"116U0520133", "1    16  U0520133", "BELC220QSKVX", "AA", "198707"},
	{"SJNVC220R00000001", "SJNVC220R00000001", "RMODEL", "AB", "198801"},
}

type partRow struct {
	etd, plate    string
	pospie        int
	mark, item    string
	variant       string
	level         int
	partNo        string
	desEN, desFR  string
	qty, capacity string
	dataplic      string
	ica, ree, pnc string
}

var parts = []partRow{
	{"AA", "AA230A", 1, "*", "01", "01", 1, "-11715-D5500", "TENSIONER", "TENDEUR", "1", "", "0487-0487", "", "", "11715"},
	{"AA", "AA230A", 2, "", "", "02", 2, "-11716-D5500", "BOLT", "BOULON", "2", "10", "0487-1294", "", "", ""},
	{"AA", "AA230A", 3, "#", "02", "01", 1, "-23319-D9700", "BEARING HOUSING", "PALIER", "2", "", "0692-1194", "2-0", "-03902204-0", "23319"},
	{"AA", "AA231", 4, "*", "01", "01", 1, "-23100-D9701", "ALTERNATOR ASSY", "ALTERNATEUR COMPLET", "1", "", "-0288", "", "", "23100"},
	{"AA", "AA101", 5, "*", "01", "01", 1, "-03902204-0", "ROLLER BEARING", "ROULEMENT", "2", "2", "0487-0692", "", "", "23358"},
	{"AA", "AA231", 6, "*", "02", "01", 1, "-23319-D9799", "BEARING HOUSING", "PALIER", "2", "", "1194-", "2-2", "-23319-D9700", "23319"},
	{"AB", "AB040", 1, "*", "01", "01", 1, "-90000-00001", "CYCLE A", "", "1", "", "0187-1294", "", "-90000-00002", ""},
	{"AB", "AB040", 2, "*", "02", "01", 1, "-90000-00002", "CYCLE B", "", "1", "", "0187-1294", "", "-90000-00001", ""},
}

var files = []string{
	"img/AA/AA230A.png", "img/AA/AA231.png", "img/AA/AA101.png", "img/AB/AB040.png",
	"gindex/AA/A.png", "gindex/AA/B.png", "cinfo/AA/G0101.pdf",
}

// NewDataDir creates a complete synthetic data directory under t.TempDir() and returns its path.
func NewDataDir(t testing.TB) string {
	t.Helper()
	dir := t.TempDir()
	write(t, filepath.Join(dir, "manifest.json"),
		`{"schema":1,"version":"test-1","edition":"Ed. TEST","built":"2026-10-06"}`)
	for _, f := range files {
		write(t, filepath.Join(dir, filepath.FromSlash(f)), "x")
	}

	db, err := sql.Open("sqlite", filepath.Join(dir, "data.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mustExec(t, db, Schema)
	mustExec(t, db, rows)
	for _, v := range vins {
		mustExec(t, db, "INSERT INTO vin VALUES (?,?,?,?,?,?,?)", v.vin, v.raw, v.model, v.etd, "4", v.prod, rev(v.vin))
	}
	for i := range parts {
		p := &parts[i]
		langs := []string{"en", "fr"}
		if p.etd == "AB" {
			langs = []string{"en"}
		}
		for _, lang := range langs {
			des := p.desEN
			if lang == "fr" {
				des = p.desFR
			}
			ind := [5]any{}
			ind[p.level-1] = "-"
			mustExec(t, db, `INSERT INTO part VALUES (?,?,?,?,?,?,?,NULL,?,?,?,?,?,?,?,?,?,NULL,?,?,NULL,NULL,?,?,?,?,NULL,?)`,
				p.etd, lang, p.pospie, p.plate, null(p.item), itemEff(p), p.variant, null(p.mark),
				ind[0], ind[1], ind[2], ind[3], ind[4],
				p.partNo, key(p.partNo), des, p.qty, null(p.capacity), p.dataplic,
				null(p.ica), null(p.ree), null(key(p.ree)), null(p.pnc))
		}
	}
	mustExec(t, db, "INSERT INTO part_fts(part_fts) VALUES ('rebuild')")
	return dir
}

// itemEff mimics the extraction: an empty item continues the previous item of the same plate.
func itemEff(p *partRow) string {
	if p.item != "" {
		return p.item
	}
	for i := len(parts) - 1; i >= 0; i-- {
		q := &parts[i]
		if q.etd == p.etd && q.plate == p.plate && q.pospie < p.pospie && q.item != "" {
			return q.item
		}
	}
	return ""
}

func key(ref string) string {
	var b strings.Builder
	for _, r := range strings.ToUpper(ref) {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func rev(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

func null(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func mustExec(t testing.TB, db *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(), query, args...); err != nil {
		t.Fatalf("fixture SQL failed: %v\n%s", err, query)
	}
}

func write(t testing.TB, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
