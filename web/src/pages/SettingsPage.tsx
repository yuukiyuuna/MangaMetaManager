import React, { useEffect, useState } from 'react';
import { Loader2 } from 'lucide-react';
import { api } from '../api';
import type { ProxySettings, AppSettings, LibraryFolder } from '../api';

import ProxySection from '../components/settings/ProxySection';
import FolderSection from '../components/settings/FolderSection';
import AppSettingsSection from '../components/settings/AppSettingsSection';

const SettingsPage: React.FC = () => {
  const [loading, setLoading] = useState(true);
  const [proxy, setProxy] = useState<ProxySettings | null>(null);
  const [appSettings, setAppSettings] = useState<AppSettings | null>(null);
  const [folders, setFolders] = useState<LibraryFolder[] | null>(null);

  const fetchAllSettings = async () => {
    setLoading(true);
    try {
      const [proxyRes, appRes, folderRes] = await Promise.all([
        api.getProxySettings(),
        api.getAppSettings(),
        api.getLibraryFolders()
      ]);
      setProxy(proxyRes.data);
      setAppSettings(appRes.data);
      setFolders(folderRes.data);
    } catch (err) {
      console.error('Failed to fetch settings:', err);
    }
    setLoading(false);
  };

  useEffect(() => {
    fetchAllSettings();
  }, []);

  if (loading || !proxy || !appSettings || !folders) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-8 h-8 animate-spin text-blue-600" />
      </div>
    );
  }

  return (
    <div className="max-w-4xl space-y-8">
      <h2 className="text-3xl font-bold text-gray-800">Settings</h2>

      <ProxySection initialData={proxy} />
      
      <FolderSection initialFolders={folders} />

      <AppSettingsSection initialData={appSettings} />
    </div>
  );
};

export default SettingsPage;
