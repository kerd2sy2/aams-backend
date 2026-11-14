package system

import "context"

type SettingService interface {
	GetAllSettings(ctx context.Context) (AppSettingsResponse, error)
	GetSettingByKey(ctx context.Context, key string) (*AppSetting, error)
	UpdateAppSettings(ctx context.Context, req UpdateAppSettingsRequest) error
	UpdateSetting(ctx context.Context, key, value string) error
}

type settingService struct {
	repo SystemRepository
}

func NewSettingService(repo SystemRepository) SettingService {
	return &settingService{repo: repo}
}

func (s *settingService) GetAllSettings(ctx context.Context) (AppSettingsResponse, error) {
	settings, err := s.repo.GetSettings(ctx)
	if err != nil {
		return AppSettingsResponse{}, err
	}

	resp := AppSettingsResponse{
		SiteName: "AAMS Logistics System",
	}

	for _, set := range settings {
		switch set.Key {
		case "site_name":
			resp.SiteName = set.Value
		case "logo_url":
			resp.LogoURL = set.Value
		case "app_download_url":
			resp.AppDownloadURL = set.Value
		}
	}
	if resp.AppDownloadURL == "" {
		resp.AppDownloadURL = "https://play.google.com/store/apps/details?id=com.aams.app"
	}
	return resp, nil
}

func (s *settingService) GetSettingByKey(ctx context.Context, key string) (*AppSetting, error) {
	return s.repo.GetSettingByKey(ctx, key)
}

func (s *settingService) UpdateAppSettings(ctx context.Context, req UpdateAppSettingsRequest) error {
	if req.SiteName != "" {
		if err := s.repo.UpsertSetting(ctx, &AppSetting{Key: "site_name", Value: req.SiteName}); err != nil {
			return err
		}
	}
	if req.LogoURL != "" {
		if err := s.repo.UpsertSetting(ctx, &AppSetting{Key: "logo_url", Value: req.LogoURL}); err != nil {
			return err
		}
	}
	if req.AppDownloadURL != "" {
		if err := s.repo.UpsertSetting(ctx, &AppSetting{Key: "app_download_url", Value: req.AppDownloadURL}); err != nil {
			return err
		}
	}
	return nil
}

func (s *settingService) UpdateSetting(ctx context.Context, key, value string) error {
	return s.repo.UpsertSetting(ctx, &AppSetting{Key: key, Value: value})
}
