package zonefile

type ExportOrigin struct {
	Origin string `json:"origin"`
}

/*
type ZoneFileServer struct {
	// key: origin
	zoneFileModels map[string]*zonefileutil.ZoneFileModel
}

func NewZoneFileServer(path string) (*ZoneFileServer, error) {
	c := &ZoneFileServer{}
	c.zoneFileModels = make(map[string]*zonefileutil.ZoneFileModel)

	belogs.Debug("NewZoneFileServer():path:", path)
	isDir, err := osutil.IsDir(path)
	if err != nil || !isDir {
		belogs.Error("NewZoneFileServer(): IsDir fail:", path, err)
		return nil, errors.New(path + " is not directory")
	}

	suffixs := make(map[string]string)
	suffixs[".zone"] = ".zone"
	files, err := osutil.GetAllFilesBySuffixs(path, suffixs)
	if err != nil {
		belogs.Error("NewZoneFileServer():GetAllFilesBySuffixs fail:", path, err)
		return nil, err
	}
	belogs.Debug("NewZoneFileServer():files:", files)
	for _, file := range files {
		zoneFileModel, err := zonefileutil.LoadZoneFile(file)
		if err != nil {
			belogs.Error("NewZoneFileServer():LoadZoneFile fail:", file, err)
			continue
		}
		belogs.Debug("NewZoneFileServer(): zoneFileModel:", jsonutil.MarshalJson(zoneFileModel))
		c.zoneFileModels[zoneFileModel.Origin] = zoneFileModel

	}
	belogs.Info("NewZoneFileServer(): zoneFileModels:", jsonutil.MarshalJson(c.zoneFileModels))
	return c, nil
}

type UpdateResourceRecord struct {
	OldResourceRecord zonefileutil.ResourceRecord `json:"oldResourceRecord"`
	NewResourceRecord zonefileutil.ResourceRecord `json:"newResourceRecord"`
}
*/
