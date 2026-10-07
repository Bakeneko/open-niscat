# Dump every user table of every .mdb under a directory to JSON Lines, faithfully.
# Jet 3 (Access 97) files are only readable through the 32-bit Jet OLEDB provider,
# so this script MUST run under C:\Windows\SysWOW64\WindowsPowerShell\v1.0\powershell.exe.
#
# Output: <OutDir>\<relative mdb path with / -> __, no ext>\<table>.jsonl
#         plus _schema.json per mdb (column names, OLEDB types, row counts).
param(
    [Parameter(Mandatory)][string]$SrcDir,
    [Parameter(Mandatory)][string]$OutDir
)
$ErrorActionPreference = 'Stop'
if ([IntPtr]::Size -ne 4) { throw "Run this script with the 32-bit PowerShell (SysWOW64)." }

Add-Type -ReferencedAssemblies System.Data -TypeDefinition @'
using System;
using System.Data;
using System.Data.OleDb;
using System.Globalization;
using System.IO;
using System.Text;

public static class MdbDump {
    static string Esc(string s) {
        var sb = new StringBuilder(s.Length + 2);
        sb.Append('"');
        foreach (char c in s) {
            switch (c) {
                case '"': sb.Append("\\\""); break;
                case '\\': sb.Append("\\\\"); break;
                case '\n': sb.Append("\\n"); break;
                case '\r': sb.Append("\\r"); break;
                case '\t': sb.Append("\\t"); break;
                default:
                    if (c < 0x20) sb.AppendFormat("\\u{0:x4}", (int)c); else sb.Append(c);
                    break;
            }
        }
        sb.Append('"');
        return sb.ToString();
    }

    static string Val(object v) {
        if (v == null || v is DBNull) return "null";
        if (v is string) return Esc((string)v);
        if (v is bool) return ((bool)v) ? "true" : "false";
        if (v is DateTime) return Esc(((DateTime)v).ToString("yyyy-MM-ddTHH:mm:ss", CultureInfo.InvariantCulture));
        if (v is byte[]) return Esc("base64:" + Convert.ToBase64String((byte[])v));
        if (v is float || v is double || v is decimal)
            return Convert.ToString(v, CultureInfo.InvariantCulture);
        if (v is IConvertible) return Convert.ToString(v, CultureInfo.InvariantCulture);
        return Esc(v.ToString());
    }

    // Returns row count.
    public static long DumpTable(OleDbConnection c, string table, string outPath) {
        var cmd = c.CreateCommand();
        cmd.CommandText = "SELECT * FROM [" + table + "]";
        long n = 0;
        using (var r = cmd.ExecuteReader())
        using (var w = new StreamWriter(outPath, false, new UTF8Encoding(false))) {
            var names = new string[r.FieldCount];
            for (int i = 0; i < r.FieldCount; i++) names[i] = Esc(r.GetName(i));
            var sb = new StringBuilder();
            while (r.Read()) {
                sb.Length = 0;
                sb.Append('{');
                for (int i = 0; i < r.FieldCount; i++) {
                    if (i > 0) sb.Append(',');
                    sb.Append(names[i]).Append(':').Append(Val(r.GetValue(i)));
                }
                sb.Append('}');
                w.Write(sb.ToString());
                w.Write('\n');
                n++;
            }
        }
        return n;
    }
}
'@

$src = (Resolve-Path $SrcDir).Path.TrimEnd('\')
New-Item -ItemType Directory -Force $OutDir | Out-Null
$out = (Resolve-Path $OutDir).Path

Get-ChildItem $src -Recurse -Filter *.mdb | Sort-Object FullName | ForEach-Object {
    $rel = $_.FullName.Substring($src.Length + 1)
    $key = ($rel -replace '\.mdb$', '') -replace '[\\/]', '__'
    $dir = Join-Path $out $key
    New-Item -ItemType Directory -Force $dir | Out-Null

    $c = New-Object System.Data.OleDb.OleDbConnection("Provider=Microsoft.Jet.OLEDB.4.0;Data Source=$($_.FullName);Mode=Read")
    $c.Open()
    $schema = [ordered]@{ source = $rel; tables = [ordered]@{} }
    $tables = $c.GetOleDbSchemaTable([System.Data.OleDb.OleDbSchemaGuid]::Tables, $null).Rows |
        Where-Object { $_.TABLE_TYPE -eq 'TABLE' } | ForEach-Object { $_.TABLE_NAME }
    foreach ($t in $tables) {
        $cols = $c.GetOleDbSchemaTable([System.Data.OleDb.OleDbSchemaGuid]::Columns, @($null, $null, $t, $null)).Rows |
            Sort-Object ORDINAL_POSITION |
            ForEach-Object { [ordered]@{ name = $_.COLUMN_NAME; type = [int]$_.DATA_TYPE; size = $_.CHARACTER_MAXIMUM_LENGTH } }
        $n = [MdbDump]::DumpTable($c, $t, (Join-Path $dir "$t.jsonl"))
        $schema.tables[$t] = [ordered]@{ rows = $n; columns = @($cols) }
    }
    $c.Close()
    $schema | ConvertTo-Json -Depth 6 | Out-File -Encoding utf8 (Join-Path $dir '_schema.json')
    Write-Output ("{0}: {1} tables" -f $rel, $tables.Count)
}
