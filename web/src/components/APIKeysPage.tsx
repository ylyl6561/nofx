import { useState, useEffect } from 'react';
import { Copy, Plus, Trash2, AlertCircle } from 'lucide-react';

interface APIKey {
  id: string;
  key_prefix: string;
  name: string;
  enabled: boolean;
  rate_limit: number;
  usage_count: number;
  last_used_at?: string;
  created_at: string;
}

export default function APIKeysPage() {
  const [apiKeys, setApiKeys] = useState<APIKey[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [newKeyName, setNewKeyName] = useState('');
  const [newKeyRateLimit, setNewKeyRateLimit] = useState(1000);
  const [createdKey, setCreatedKey] = useState<string | null>(null);
  const [usageStats, setUsageStats] = useState<any>(null);

  useEffect(() => {
    fetchAPIKeys();
    fetchUsageStats();
  }, []);

  const fetchAPIKeys = async () => {
    try {
      const token = localStorage.getItem('token');
      const response = await fetch('/api/api-keys', {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });
      const data = await response.json();
      setApiKeys(data.keys || []);
    } catch (error) {
      console.error('Failed to fetch API keys:', error);
    } finally {
      setLoading(false);
    }
  };

  const fetchUsageStats = async () => {
    try {
      const token = localStorage.getItem('token');
      const response = await fetch('/api/api-keys/usage', {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });
      const data = await response.json();
      setUsageStats(data);
    } catch (error) {
      console.error('Failed to fetch usage stats:', error);
    }
  };

  const createAPIKey = async () => {
    try {
      const token = localStorage.getItem('token');
      const response = await fetch('/api/api-keys', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          name: newKeyName,
          rate_limit: newKeyRateLimit,
        }),
      });
      const data = await response.json();
      
      if (response.ok) {
        setCreatedKey(data.api_key);
        setNewKeyName('');
        setNewKeyRateLimit(1000);
        fetchAPIKeys();
      } else {
        alert(`Failed to create API key: ${data.message}`);
      }
    } catch (error) {
      console.error('Failed to create API key:', error);
      alert('Failed to create API key');
    }
  };

  const deleteAPIKey = async (keyId: string) => {
    if (!confirm('Are you sure you want to delete this API key?')) {
      return;
    }

    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`/api/api-keys/${keyId}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (response.ok) {
        fetchAPIKeys();
      } else {
        alert('Failed to delete API key');
      }
    } catch (error) {
      console.error('Failed to delete API key:', error);
      alert('Failed to delete API key');
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    alert('Copied to clipboard!');
  };

  if (loading) {
    return <div className="p-8">Loading...</div>;
  }

  return (
    <div className="p-8 max-w-6xl mx-auto">
      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-2">API Keys</h1>
        <p className="text-gray-600">
          Manage your API keys to access NOFX trading services programmatically
        </p>
      </div>

      {/* Usage Stats */}
      {usageStats && (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
          <div className="bg-white p-6 rounded-lg shadow">
            <div className="text-sm text-gray-600 mb-1">Total API Calls</div>
            <div className="text-2xl font-bold">{usageStats.total_calls}</div>
          </div>
          <div className="bg-white p-6 rounded-lg shadow">
            <div className="text-sm text-gray-600 mb-1">Monthly Usage</div>
            <div className="text-2xl font-bold">{usageStats.monthly_usage}</div>
          </div>
          <div className="bg-white p-6 rounded-lg shadow">
            <div className="text-sm text-gray-600 mb-1">Avg Response Time</div>
            <div className="text-2xl font-bold">
              {usageStats.avg_response_time?.toFixed(0) || 0}ms
            </div>
          </div>
        </div>
      )}

      {/* Create API Key Button */}
      <div className="mb-6">
        <button
          onClick={() => setShowCreateModal(true)}
          className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
        >
          <Plus size={20} />
          Create New API Key
        </button>
      </div>

      {/* API Keys List */}
      <div className="bg-white rounded-lg shadow overflow-hidden">
        <table className="w-full">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                Name
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                Key
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                Usage
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                Last Used
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                Actions
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {apiKeys.map((key) => (
              <tr key={key.id}>
                <td className="px-6 py-4 whitespace-nowrap">
                  <div className="font-medium">{key.name}</div>
                  {!key.enabled && (
                    <span className="text-xs text-red-600">Disabled</span>
                  )}
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <code className="text-sm bg-gray-100 px-2 py-1 rounded">
                    {key.key_prefix}
                  </code>
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <div className="text-sm">
                    {key.usage_count} / {key.rate_limit}
                  </div>
                  <div className="w-full bg-gray-200 rounded-full h-2 mt-1">
                    <div
                      className="bg-blue-600 h-2 rounded-full"
                      style={{
                        width: `${Math.min((key.usage_count / key.rate_limit) * 100, 100)}%`,
                      }}
                    />
                  </div>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                  {key.last_used_at
                    ? new Date(key.last_used_at).toLocaleString()
                    : 'Never'}
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <button
                    onClick={() => deleteAPIKey(key.id)}
                    className="text-red-600 hover:text-red-800"
                  >
                    <Trash2 size={18} />
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>

        {apiKeys.length === 0 && (
          <div className="text-center py-12 text-gray-500">
            No API keys yet. Create one to get started!
          </div>
        )}
      </div>

      {/* Create API Key Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-8 max-w-md w-full">
            <h2 className="text-2xl font-bold mb-4">Create New API Key</h2>
            
            <div className="mb-4">
              <label className="block text-sm font-medium mb-2">Name</label>
              <input
                type="text"
                value={newKeyName}
                onChange={(e) => setNewKeyName(e.target.value)}
                className="w-full px-3 py-2 border rounded-lg"
                placeholder="My API Key"
              />
            </div>

            <div className="mb-6">
              <label className="block text-sm font-medium mb-2">
                Rate Limit (calls/month)
              </label>
              <input
                type="number"
                value={newKeyRateLimit}
                onChange={(e) => setNewKeyRateLimit(parseInt(e.target.value))}
                className="w-full px-3 py-2 border rounded-lg"
              />
            </div>

            <div className="flex gap-3">
              <button
                onClick={createAPIKey}
                className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
              >
                Create
              </button>
              <button
                onClick={() => setShowCreateModal(false)}
                className="flex-1 px-4 py-2 bg-gray-200 rounded-lg hover:bg-gray-300"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Show Created Key Modal */}
      {createdKey && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-8 max-w-2xl w-full">
            <div className="flex items-center gap-2 text-yellow-600 mb-4">
              <AlertCircle size={24} />
              <h2 className="text-2xl font-bold">Save Your API Key</h2>
            </div>
            
            <p className="text-gray-600 mb-4">
              This is the only time you'll see this key. Please save it securely!
            </p>

            <div className="bg-gray-100 p-4 rounded-lg mb-4 flex items-center justify-between">
              <code className="text-sm break-all">{createdKey}</code>
              <button
                onClick={() => copyToClipboard(createdKey)}
                className="ml-4 p-2 hover:bg-gray-200 rounded"
              >
                <Copy size={20} />
              </button>
            </div>

            <button
              onClick={() => {
                setCreatedKey(null);
                setShowCreateModal(false);
              }}
              className="w-full px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
            >
              I've Saved My Key
            </button>
          </div>
        </div>
      )}

      {/* API Documentation */}
      <div className="mt-8 bg-gray-50 p-6 rounded-lg">
        <h3 className="text-lg font-bold mb-4">Quick Start</h3>
        <div className="space-y-4">
          <div>
            <div className="text-sm font-medium mb-2">1. List your traders:</div>
            <pre className="bg-gray-800 text-green-400 p-4 rounded overflow-x-auto text-sm">
{`curl -X GET https://your-domain.com/api/v1/traders \\
  -H "Authorization: Bearer aitrader_your_api_key"`}
            </pre>
          </div>

          <div>
            <div className="text-sm font-medium mb-2">2. Start a trader:</div>
            <pre className="bg-gray-800 text-green-400 p-4 rounded overflow-x-auto text-sm">
{`curl -X POST https://your-domain.com/api/v1/traders/{id}/start \\
  -H "Authorization: Bearer aitrader_your_api_key" \\
  -H "Content-Type: application/json" \\
  -d '{"trader_id": "your_trader_id"}'`}
            </pre>
          </div>
        </div>
      </div>
    </div>
  );
}
