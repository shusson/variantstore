CREATE DATABASE variants;
USE variants;

CREATE TABLE vs (
		VARIANT TEXT,
		CHROMOSOME VARCHAR(2),
		START INT,
		REF TEXT,
		ALT TEXT,
		RSID TEXT,
		AC INT,
		AF DECIMAL(6,5),
    nHomRef INT,
    nHet INT,
    nHomVar INT,
		TYPE VARCHAR(32),
		CATO DECIMAL(6,5),
		eigen DECIMAL(6,5),
		sift TEXT,
		polyPhen TEXT,
		tgpAF DECIMAL(6,5),
		hrcAF DECIMAL(6,5),
		gnomadAF DECIMAL(6,5),
		gnomadAF_AFR DECIMAL(6,5),
		gnomadAF_AMR DECIMAL(6,5),
		gnomadAF_ASJ DECIMAL(6,5),
		gnomadAF_EAS DECIMAL(6,5),
		gnomadAF_FIN DECIMAL(6,5),
		gnomadAF_NFE DECIMAL(6,5),
		gnomadAF_OTHD DECIMAL(6,5),
		ensemblId TEXT,
		consequences TEXT,
		geneSymbol TEXT,
		clinvar TEXT,
		wasSplit TEXT,
		INDEX name (chromosome, start)
	);

LOAD DATA INFILE '/data/mgrb.tsv' INTO TABLE vs
				  FIELDS TERMINATED BY '\t'
				  LINES TERMINATED BY '\n'
				  IGNORE 1 LINES;