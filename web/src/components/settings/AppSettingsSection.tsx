import React, { useState } from 'react';
import { api } from '../../api';
import type { AppSettings } from '../../api';
import { showToast } from '../Toast';

interface AppSettingsSectionProps {
  initialData: AppSettings;
}

const AppSettingsSection: React.FC<AppSettingsSectionProps> = ({ initialData }) => {
  const [appSettings, setAppSettings] = useState<AppSettings>(initialData);

  const saveAppSettings = async (newSettings: AppSettings) => {
    try {
      await api.updateAppSettings(newSettings);
      setAppSettings(newSettings);
      showToast('App settings updated!');
    } catch (err) {
      console.error(err);
      showToast('Failed to update app settings.', 'error');
    }
  };

  return (
    <section className="bg-white p-6 rounded-xl shadow-sm border border-gray-100">
      <h3 className="text-xl font-semibold text-gray-800 mb-6">App Settings</h3>
      
      <div className="space-y-6">
        <div className="flex items-start justify-between">
          <div className="max-w-2xl">
            <label className="text-lg font-medium text-gray-800">
              Backup Before Metadata Rewrite (写入元数据前备份)
            </label>
            <p className="text-gray-500 mt-1">
              When saving metadata, the app restructures (flattens) the ZIP/CBZ structure to ensure 100% compatibility with Komga. 
              Enabling this will create a <code className="bg-gray-100 px-1 rounded">.bak</code> copy of your original file before the first modification.
              <br />
              <span className="text-sm italic">
                写入元数据时为了兼容 Komga 会展平压缩包结构。开启此项后，程序会在修改前生成一份 .bak 备份，以防万一。
              </span>
            </p>
          </div>
          <button
            onClick={() => saveAppSettings({ ...appSettings, backupBeforeFlatten: !appSettings.backupBeforeFlatten })}
            className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors shrink-0 mt-1 ${
              appSettings.backupBeforeFlatten ? 'bg-blue-600' : 'bg-gray-200'
            }`}
          >
            <span className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
              appSettings.backupBeforeFlatten ? 'translate-x-6' : 'translate-x-1'
            }`} />
          </button>
        </div>
      </div>
    </section>
  );
};

export default AppSettingsSection;
