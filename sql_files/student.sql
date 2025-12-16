WITH CTE_GRID_DOCUMENT AS 
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
	SELECT 
		Student.[DBID], 
		Student.Stud_ID, 
		Student.Last_Name, 
		Student.First_Name, 
		Student.Locked, 
		Student.Mi, 
		Student.Local_ID, 
		Student.Aid_Eligible, 
		Student.Aide_Req, 
		Student.Comments, 
		Student.[Disabled], 
		district.district AS District, 
		Student.Dob, 
		Student.Entry_Date, 
		Student.Geo_City, 
		Student.Geo_County, 
		Student.Geo_Street, 
		Grade.Code AS grade, 
		MailingCity.[Name] AS Mail_City, 
		MailingState.[NAME] AS Mail_State, 
		Student.Mail_Street1, 
		Student.Mail_Street2, 
		MailingPostalCode.Postal AS Mail_Zip, 
		Student.DistanceFromSchl, 
		Student.School, 
		Gender.Code AS Sex, -- Forward Compatibility. 
		Gender.Code AS Gender, 
		Student.Transported, 
		Student.XCoord, 
		Student.YCoord, 
		Student.Geo_Zip, 
		Student.GeoConfidence, 
		Student.IntGratNum1, 
		Student.IntGratNum2, 
		Student.IntGratChar1, 
		Student.IntGratChar2, 
		Student.IntGratDate1, 
		Student.IntGratDate2, 
		Student.PreRedistSchool, 
		Student.CreatedOn, 
		Student.CreatedBy, 
		ISNULL(CreatedByUser.LoginId, 'Public') AS CreatedByName, 
		Student.LastUpdated, 
		Student.LastUpdatedID, 
		ISNULL([User].LoginId, 'Public') AS LastUpdatedName, 
		Student.LastUpdatedType, 
		Student.[GUID], 
		Student.PriorSchool, 
		Student.ProhibitCross, 
		Student.Cohort, 
		Student.DistanceFromAMStop, 
		Student.DistanceFromPMStop, 
		Student.InActive, 
		Student.LoadTime, 
		Student.LoadTimeManuallyChanged, 
		Student.ResidSchool, 
		Student.DistanceFromResidSch, 
		StudentTag.TagId, 
		Student.System_Street, 
		Student.System_City, 
		Student.System_State, 
		Student.System_Zip, 
		Student.Email, 
		fromSR.PuLocation AS FromSchoolPUDefaultRequirement, 
		fromSR.DoLocation AS FromSchoolDODefaultRequirement, 
		toSR.PuLocation AS ToSchoolPUDefaultRequirement, 
		toSR.DoLocation AS ToSchoolDODefaultRequirement, 
		CASE WHEN ISNULL(last_name, '') != '' AND ISNULL(first_name, '') != '' THEN last_name + ', ' + first_name 
			WHEN ISNULL(last_name, '') != '' THEN last_name 
			WHEN ISNULL(first_name, '') != '' THEN first_name 
			ELSE '' END  AS FullName, 
		IIF(student.XCOORD <> 0,'4', '') AS GEO, 
		school.[Name] AS SCHOOLNAME, 
		ressch.[Name] AS ResSchName, 
		IIF(NOT (student.Comments IS NULL OR student.Comments LIKE ''), '5', '') AS Notes, 
		IIF(ISNUMERIC(LTRIM(RTRIM(SUBSTRING([student].[mail_Street1], 1, CHARINDEX(' ', [student].[mail_Street1] + ' '))))) = 1, LTRIM(RTRIM(SUBSTRING([student].[mail_Street1], 1, CHARINDEX(' ', [student].[mail_Street1] + ' ')))), '') AS MailStreetNumber, 
		IIF(ISNUMERIC(LTRIM(RTRIM(SUBSTRING([student].[mail_Street1], 1, CHARINDEX(' ', [student].[mail_Street1] + ' '))))) = 1, LTRIM(RTRIM(SUBSTRING(student.mail_Street1, CHARINDEX(' ', student.mail_Street1 + ' '),LEN(student.mail_Street1) - CHARINDEX(' ', student.mail_Street1 + ' ') + 1))), student.mail_street1) AS MailStreetName, 
		IIF(ISNUMERIC(LTRIM(RTRIM(SUBSTRING([student].[geo_street], 1, CHARINDEX(' ', [student].[geo_street] + ' '))))) = 1, LTRIM(RTRIM(SUBSTRING([student].[geo_street], 1, CHARINDEX(' ', [student].[geo_street] + ' ')))), '') AS GeoStreetNumber, 
		IIF(ISNUMERIC(LTRIM(RTRIM(SUBSTRING([student].[geo_street],1,CHARINDEX(' ',[student].[geo_street]+' '))))) = 1,LTRIM(RTRIM(SUBSTRING(student.geo_street,CHARINDEX(' ',student.geo_street + ' '),LEN(student.geo_street)-CHARINDEX(' ',student.geo_street + ' ')+1))),student.geo_street)  AS GeoStreetName,                   
		IIF(student.intgratnum1 = 0, '', IIF(student.intgratnum1 = 1, 'Add', 'Update')) AS IntegrationAction, 
		IIF(student.intgratnum1 >= 100, 'Y', '') AS IntegrationUnRoute, IIF(student.intgratnum1 >= 1000, 'Y', '') AS IntegrationUnGeoCode, 
		EthnicCodes.names AS EthnicCode, 
		IIF(CTE_districtstudentpolicy.Value1 > 0,CTE_districtstudentpolicy.Value1, ' ') AS WalkToStopPolicy, 
		IIF(CTE_districtstudentpolicy.Value2 > 0, CTE_districtstudentpolicy.Value2, ' ') AS WalkToSchoolPolicy, 
		IIF(CTE_districtstudentpolicy.Value4 > 0, dbo.fnVal(CTE_districtstudentpolicy.Value4), ' ') AS RideTimePolicy, 
		IIF(DATEPART(MONTH,[dob]) > DATEPART(MONTH, GETDATE()), DATEDIFF(YEAR, [dob], GETDATE()) - 1, 
				IIF(DATEPART(MONTH, [dob]) = DATEPART(MONTH, GETDATE()) AND DATEPART(DAY, [dob]) > DATEPART(DAY,GETDATE()), 
				DATEDIFF(YEAR, [dob], GETDATE()) - 1, 
				DATEDIFF(YEAR, [dob], GETDATE()))) AS AGE, 
		0 AS Selected, 
		PrimaryContact.PrimaryContactName, 
		PrimaryContact.PrimaryContactPhone, 
		PrimaryContact.PrimaryContactMobile, 
		PrimaryContact.PrimaryContactEmail, 
		vwStudentLoadTime.ActualLoadTime, 
		vwStudentLoadTime.CalculatedLoadTime, 
		ISNULL(sfc.StopfinderContacts, 0) AS StopfinderContacts, 
		DisabilityCodes.DisabilityCodes, 
		ISNULL(drs.DocumentCount, 0) AS DocumentCount, 
		POPULATIONREGION.[Name] AS PopulationRegionName, 
		Student.PopulationRegionID, 
		Student.LastUngeocoded, 
		CASE WHEN Student.LastUngeocodedReason = 1 THEN 'Import' 
		WHEN Student.LastUngeocodedReason = 2 THEN 'Manual Address Change' 
		WHEN Student.LastUngeocodedReason = 3 THEN 'Mass Update' 
		WHEN Student.LastUngeocodedReason = 4 THEN 'Ungeocode Tool' 
		WHEN Student.LastUngeocodedReason = 5 THEN 'Form Update' 
		ELSE '' END  AS LastUngeocodedReason, 
		vtr.Tags, 
		Student.ID,
		rp.ID as RecordPictureID
	FROM dbo.Student 
	LEFT JOIN dbo.RecordPicture rp
	ON rp.DBID = Student.DBID and rp.RecordID = Student.Stud_ID and rp.DataType = 9 
	LEFT JOIN dbo.MailingCity 
	ON MailingCity.ID = Student.mail_city 
	LEFT JOIN dbo.MailingPostalCode 
	ON MailingPostalCode.ID = Student.mail_zip 
	LEFT JOIN dbo.MailingState 
	ON MailingState.ID = Student.Mail_State_Id 
	LEFT JOIN dbo.grade 
	ON grade.ID = student.grade 
	LEFT JOIN dbo.Gender 
	ON Gender.ID = student.Gender 
	LEFT JOIN POPULATIONREGION 
	ON POPULATIONREGION.OBJECTID = Student.PopulationRegionID 
	AND POPULATIONREGION.[DBID] = Student.[DBID] 
	LEFT JOIN PrimaryContact 
	ON PrimaryContact.RecordId = student.Stud_ID 
	AND PrimaryContact.[DBID] = student.[DBID] 
	CROSS APPLY dbo.fnCurrentDateByTimezone_iTVF() AS tz 
	outer apply  
	( 
		select top 1 TagId from StudentTagId where [DBID]=student.[DBID]  
		and StudentId = student.Stud_ID  
		and tz.CurrentDateTime BETWEEN StartDate AND ISNULL(EndDate,CAST('9999/12/31 23:59:59' AS DATETIME))        
		order by StartDate DESC, LastUpdated DESC, ID DESC 
	) as StudentTag 
	LEFT JOIN dbo.[vwStudentRequirement] fromSR 
	ON fromSR.Stud_ID = Student.Stud_ID 
	AND fromSR.[DBID] = Student.[DBID] 
	AND fromSR.[Type] = 0 
	AND fromSR.[Session] = 'From School' 
	LEFT JOIN dbo.[vwStudentRequirement] toSR 
	ON toSR.Stud_ID = Student.Stud_ID 
	AND toSR.[DBID] = Student.[DBID] 
	AND toSR.[Type] = 0 
	AND toSR.[Session] = 'To School' 
	LEFT JOIN school 
	ON STUDENT.school = school.SchoolCode 
	AND school.[DBID] = Student.[DBID] 
	LEFT JOIN school ressch 
	ON STUDENT.ResidSchool = ressch.SchoolCode 
	AND ressch.[DBID] = Student.[DBID] 
	LEFT JOIN district 
	ON STUDENT.DistrictID = district.DistrictID 
	AND district.[DBID] = Student.[DBID] 
	LEFT JOIN CTE_districtstudentpolicy 
	ON grade.code = CTE_districtstudentpolicy.GradeCode 
	AND Student.[DBID] = CTE_districtstudentpolicy.[DBID] 
	LEFT JOIN dbo.vwStudentLoadTime 
	ON vwstudentloadtime.stud_id = Student.stud_id 
	AND vwstudentloadtime.[DBID] = Student.[DBID] 
	LEFT JOIN [User] 
	ON [User].UserID = Student.LastUpdatedId 
	LEFT JOIN [User] AS CreatedByUser 
	ON CreatedByUser.UserID = Student.CreatedBy 
	LEFT JOIN EthnicCodes 
	ON EthnicCodes.[DBID] = Student.[DBID] 
	AND EthnicCodes.Stud_ID = Student.Stud_ID 
	LEFT JOIN DisabilityCodes 
	ON DisabilityCodes.[DBID] = Student.[DBID] 
	AND DisabilityCodes.Stud_ID = Student.Stud_ID 
	LEFT JOIN sfc 
	ON Student.[DBID] = sfc.[DBID] 
	AND Student.Stud_ID = sfc.RecordID 
	LEFT JOIN drs 
	ON Student.[DBID] = drs.[DBID] 
	AND Student.Stud_ID = drs.AttachedToID 
	LEFT JOIN [vwTagRelationship] vtr ON vtr.AttachedToId = Student.Stud_ID and vtr.DBID = Student.DBID and vtr.AttachedToType = 9