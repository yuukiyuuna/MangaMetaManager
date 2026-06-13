import React, { useState } from 'react';
import { Save, CheckCircle, XCircle } from 'lucide-react';
import { api } from '../../api';
import type { ProxySettings } from '../../api';
import { showToast } from '../Toast';

interface ProxySectionProps {
  initialData: ProxySettings;
}

const ProxySection: React.FC<ProxySectionProps> = ({ initialData }) => {
  const [proxy, setProxy] = useState<ProxySettings>(initialData);
  const [testResult, setTestResult] = useState<{ success: boolean; msg: string } | null>(null);

  const saveProxy = async () => {
    try {
      await api.updateProxySettings(proxy);
      showToast('Proxy settings saved successfully!');
    } catch (err) {
      console.error(err);
      showToast('Failed to save proxy settings.', 'error');
    }
  };

  const testProxy = async () => {
    setTestResult(null);
    try {
      const res = await api.testProxy();
      if (res.data.success) {
        setTestResult({ success: true, msg: 'Connected successfully!' });
      } else {
        setTestResult({ success: false, msg: (res.data as any).error || 'Failed to connect.' });
      }
    } catch (err) {
      setTestResult({ success: false, msg: 'Error testing proxy.' });
    }
  };

  return (
    <section className="bg-white p-6 rounded-xl shadow-sm border border-gray-100">
      <div className="flex justify-between items-center mb-6">
        <h3 className="text-xl font-semibold text-gray-800">Network Proxy</h3>
        <div className="flex items-center">
          <span className="mr-3 text-sm text-gray-500">{proxy.enabled ? 'Enabled' : 'Disabled'}</span>
          <button
            onClick={() => setProxy({ ...proxy, enabled: !proxy.enabled })}
            className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
              proxy.enabled ? 'bg-blue-600' : 'bg-gray-200'
            }`}
          >
            <span className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
              proxy.enabled ? 'translate-x-6' : 'translate-x-1'
            }`} />
          </button>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Type</label>
          <select
            value={proxy.type}
            onChange={(e) => setProxy({ ...proxy, type: e.target.value })}
            className="w-full border-gray-300 rounded-lg shadow-sm focus:ring-blue-500 focus:border-blue-500 border p-2"
          >
            <option value="http">HTTP</option>
            <option value="https">HTTPS</option>
            <option value="socks5">SOCKS5</option>
          </select>
        </div>
        <div className="md:col-span-1 flex gap-4">
          <div className="flex-1">
            <label className="block text-sm font-medium text-gray-700 mb-1">Host</label>
            <input
              type="text"
              value={proxy.host}
              onChange={(e) => setProxy({ ...proxy, host: e.target.value })}
              className="w-full border-gray-300 rounded-lg shadow-sm focus:ring-blue-500 focus:border-blue-500 border p-2"
              placeholder="e.g. 127.0.0.1"
            />
          </div>
          <div className="w-24">
            <label className="block text-sm font-medium text-gray-700 mb-1">Port</label>
            <input
              type="number"
              value={proxy.port}
              onChange={(e) => setProxy({ ...proxy, port: parseInt(e.target.value) || 0 })}
              className="w-full border-gray-300 rounded-lg shadow-sm focus:ring-blue-500 focus:border-blue-500 border p-2"
              placeholder="7890"
            />
          </div>
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Timeout (Seconds)</label>
          <input
            type="number"
            value={proxy.timeoutSeconds}
            onChange={(e) => setProxy({ ...proxy, timeoutSeconds: parseInt(e.target.value) || 30 })}
            className="w-full border-gray-300 rounded-lg shadow-sm focus:ring-blue-500 focus:border-blue-500 border p-2"
            placeholder="30"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Username (Optional)</label>
          <input
            type="text"
            value={proxy.username}
            onChange={(e) => setProxy({ ...proxy, username: e.target.value })}
            className="w-full border-gray-300 rounded-lg shadow-sm focus:ring-blue-500 focus:border-blue-500 border p-2"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Password (Optional)</label>
          <input
            type="password"
            value={proxy.password || ''}
            onChange={(e) => setProxy({ ...proxy, password: e.target.value })}
            className="w-full border-gray-300 rounded-lg shadow-sm focus:ring-blue-500 focus:border-blue-500 border p-2"
            placeholder="••••••••"
          />
        </div>
        <div className="md:col-span-2">
          <label className="block text-sm font-medium text-gray-700 mb-1">No Proxy (Bypass)</label>
          <input
            type="text"
            value={proxy.noProxy}
            onChange={(e) => setProxy({ ...proxy, noProxy: e.target.value })}
            className="w-full border-gray-300 rounded-lg shadow-sm focus:ring-blue-500 focus:border-blue-500 border p-2"
            placeholder="localhost, 127.0.0.1, .local"
          />
        </div>
      </div>

      <div className="mt-6 flex items-center gap-4">
        <button
          onClick={saveProxy}
          className="bg-blue-600 text-white px-6 py-2 rounded-lg flex items-center hover:bg-blue-700 transition-colors"
        >
          <Save className="w-4 h-4 mr-2" /> Save Proxy
        </button>
        <button
          onClick={testProxy}
          className="bg-white text-gray-700 border border-gray-300 px-6 py-2 rounded-lg hover:bg-gray-50 transition-colors"
        >
          Test Connection
        </button>
        {testResult && (
          <div className={`flex items-center text-sm ${testResult.success ? 'text-green-600' : 'text-red-600'}`}>
            {testResult.success ? <CheckCircle className="w-4 h-4 mr-1" /> : <XCircle className="w-4 h-4 mr-1" />}
            {testResult.msg}
          </div>
        )}
      </div>
    </section>
  );
};

export default ProxySection;
