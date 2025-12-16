package config

const (
	StudentReportSqlHead = `WITH CTE_GRID_DOCUMENT AS 
	( 
      SELECT 
            dr.DBID, dr.AttachedToID, dr.DocumentID, dr.AttachedToType 
      FROM 
            DocumentRelationship dr 
      JOIN  
            Document d 
      ON 
            dr.DBID = d.DBID and dr.Attachedtotype = 9 and dr.DocumentID = d.DocumentID and d.DocumentClassificationID is not null 
	), 
	PrimaryContact AS ( 
		SELECT 
			rc.RecordID, 
			rc.[DBID], 
			CASE 
				WHEN c.LastName IS NULL OR c.LastName = '' THEN c.FirstName 
				WHEN c.FirstName IS NULL OR c.FirstName = '' THEN c.LastName 
				ELSE CONCAT(c.LastName, ', ', c.FirstName) 
			END AS [PrimaryContactName], 
			c.Phone AS [PrimaryContactPhone], 
			c.Mobile AS [PrimaryContactMobile], 
			c.Email AS [PrimaryContactEmail] 
		FROM RecordContact rc 
		INNER JOIN CONTACT C 
		ON rc.ContactID = c.ID 
		WHERE rc.IsPrimary = 1), 
	EthnicCodes AS ( 
		SELECT 
			oec.stud_id, 
			ec.[dbid], 
			STRING_AGG(CAST(ec.Code AS VARCHAR(MAX)), ', ') AS names 
		FROM dbo.Ethnic_Codes ec 
		INNER JOIN dbo.OwnedEthnicCodes oec 
		ON ec.EthnicCodeID = oec.EthnicCodeID 
		AND ec.[DBID] = oec.[DBID] 
		GROUP BY ec.[dbid], oec.stud_id), 
	DisabilityCodes AS ( 
		SELECT 
			od.stud_id, 
			od.[dbid], 
			STRING_AGG(CAST(dc.Code AS VARCHAR(MAX)), ', ')WITHIN GROUP (ORDER BY dc.code) AS DisabilityCodes 
		FROM dbo.OwnedDisability od 
		INNER JOIN dbo.disability_codes dc 
		ON od.DisCodeID = dc.DisCodeID 
		AND od.[DBID] = dc.[DBID] 
		GROUP BY od.[dbid], od.stud_id), 
	sfc AS ( 
		SELECT 
			COUNT(1) AS StopfinderContacts, 
			RecordContact.[DBID], 
			RecordContact.RecordID 
		FROM dbo.RecordContact 
		WHERE RecordContact.DataType = 9 -- 9:Student 
		AND RecordContact.IsStopfinder = 1 
		GROUP BY [DBID], RecordID), 
	drs AS ( 
		SELECT 
			drs.[DBID], 
			drs.AttachedToID, 
			COUNT(drs.DocumentID) AS DocumentCount 
		FROM CTE_GRID_DOCUMENT drs 
		WHERE drs.AttachedToType = 9 -- 9:Student 
		GROUP BY drs.[DBID], drs.AttachedToID), 
	CTE_districtstudentpolicy AS ( 
		SELECT 
			d.[DBID], 
			g.Code AS GradeCode, 
			d.Value1, 
			d.Value2, 
			d.Value4 
		FROM dbo.districtstudentpolicy d 
		INNER JOIN dbo.grade g 
		ON d.GradeID = g.ID) 
		`
)