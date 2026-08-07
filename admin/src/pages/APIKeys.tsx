import { useState, useEffect } from 'react';

interface APIKey {
  id: string;
  name: string;
  key_prefix: string;
  scopes: string[];
  type: string;
  status: string;
  created_at: string;
}

const INITIAL_KEYS: APIKey[] = [
  { id: 'key_1', name: 'Development Key', key_prefix: 'sk_test_51O', scopes: ['tenant.read', 'users.read'], type: 'test', status: 'active', created_at: '2026-03-01T10:00:00Z' },
  { id: 'key_2', name: 'Production Server Key', key_prefix: 'sk_live_89A', scopes: ['tenant.read', 'tenant.write', 'users.read'], type: 'live', status: 'active', created_at: '2026-04-12T14:30:00Z' }
];

export default function APIKeys() {
  const [apiKeys, setApiKeys] = useState<APIKey[]>(INITIAL_KEYS);
  const [loading, setLoading] = useState(true);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [newName, setNewName] = useState('');
  const [newType, setNewType] = useState('test');
  const [newScopes, setNewScopes] = useState('tenant.read,tenant.write');

  useEffect(() => {
    const timer = setTimeout(() => setLoading(false), 400);
    return () => clearTimeout(timer);
  }, []);

  const handleCreateKey = (e: React.FormEvent) => {
    e.preventDefault();
    if (!newName) return;
    const scopesArr = newScopes.split(',').map(s => s.trim()).filter(Boolean);
    const prefix = newType === 'live' ? 'sk_live_' : 'sk_test_';
    const randSuffix = Math.random().toString(36).substring(2, 5).toUpperCase();
    const newKey: APIKey = {
      id: `key_${Date.now()}`,
      name: newName,
      key_prefix: `${prefix}${randSuffix}`,
      scopes: scopesArr,
      type: newType,
      status: 'active',
      created_at: new Date().toISOString()
    };
    setApiKeys([newKey, ...apiKeys]);
    setNewName('');
    setNewType('test');
    setNewScopes('tenant.read,tenant.write');
    setShowCreateModal(false);
  };

  const handleRevokeKey = (id: string) => {
    if (window.confirm('Are you sure you want to revoke this API key? This action is irreversible.')) {
      setApiKeys(apiKeys.map(k => {
        if (k.id === id) {
          return { ...k, status: 'revoked' };
        }
        return k;
      }));
    }
  };

  return (
    <div className="animate-fade-in">
      <div className="sm:flex sm:items-center sm:justify-between mb-8">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 tracking-tight">API Keys</h1>
          <p className="mt-2 text-gray-600">
            Manage API keys for programmatic access to the SaaSKit API.
          </p>
        </div>
        <div className="mt-4 sm:mt-0">
          <button
            type="button"
            onClick={() => setShowCreateModal(true)}
            className="inline-flex items-center px-4 py-2.5 border border-transparent rounded-xl text-sm font-semibold text-white bg-primary-600 hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 transition-all duration-200 shadow-sm cursor-pointer"
          >
            <svg className="w-5 h-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2.5} d="M12 4v16m8-8H4" />
            </svg>
            Create API Key
          </button>
        </div>
      </div>

      <div className="bg-white rounded-2xl shadow-sm border border-gray-100 overflow-hidden">
        <table className="min-w-full divide-y divide-gray-100">
          <thead className="bg-gray-50/50">
            <tr>
              <th scope="col" className="px-6 py-4 text-left text-xs font-semibold text-gray-500 uppercase tracking-wider">
                Key Details
              </th>
              <th scope="col" className="px-6 py-4 text-left text-xs font-semibold text-gray-500 uppercase tracking-wider">
                Scopes
              </th>
              <th scope="col" className="px-6 py-4 text-left text-xs font-semibold text-gray-500 uppercase tracking-wider">
                Type
              </th>
              <th scope="col" className="px-6 py-4 text-left text-xs font-semibold text-gray-500 uppercase tracking-wider">
                Status
              </th>
              <th scope="col" className="px-6 py-4 text-left text-xs font-semibold text-gray-500 uppercase tracking-wider">
                Created At
              </th>
              <th scope="col" className="px-6 py-4 text-right text-xs font-semibold text-gray-500 uppercase tracking-wider">
                Actions
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100 bg-white">
            {loading ? (
              <tr>
                <td colSpan={6} className="px-6 py-16 text-center text-sm text-gray-500">
                  <div className="flex items-center justify-center space-x-3">
                    <svg className="animate-spin h-5 w-5 text-primary-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    <span className="font-medium text-gray-600">Loading API keys...</span>
                  </div>
                </td>
              </tr>
            ) : apiKeys.length === 0 ? (
              <tr>
                <td colSpan={6} className="px-6 py-16 text-center text-sm text-gray-400">
                  <svg className="mx-auto h-12 w-12 text-gray-300 mb-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
                  </svg>
                  <p className="font-medium">No API keys found</p>
                </td>
              </tr>
            ) : (
              apiKeys.map((key) => (
                <tr key={key.id} className="hover:bg-gray-50/55 transition-colors duration-150">
                  <td className="px-6 py-4.5 whitespace-nowrap">
                    <div className="flex items-center">
                      <div className="flex-shrink-0 h-10 w-10">
                        <div className="h-10 w-10 rounded-xl bg-amber-50 border border-amber-100 flex items-center justify-center">
                          <svg className="w-5 h-5 text-amber-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
                          </svg>
                        </div>
                      </div>
                      <div className="ml-4">
                        <div className="text-sm font-semibold text-gray-900">{key.name}</div>
                        <div className="text-sm text-gray-500 font-mono text-xs select-all bg-gray-50 border border-gray-100 px-1.5 py-0.5 rounded-md inline-block">{key.key_prefix}••••••••</div>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4.5">
                    <div className="flex flex-wrap gap-1 max-w-xs">
                      {key.scopes.map((scope) => (
                        <span key={scope} className="inline-flex items-center px-2 py-0.5 rounded-md text-xs font-medium bg-gray-100 text-gray-800 border border-gray-200">
                          {scope}
                        </span>
                      ))}
                    </div>
                  </td>
                  <td className="px-6 py-4.5 whitespace-nowrap text-sm text-gray-500 font-medium">
                    <span className={`inline-flex rounded-full px-2.5 py-0.5 text-xs font-semibold border ${
                      key.type === 'live' 
                        ? 'bg-rose-50 text-rose-700 border-rose-200' 
                        : 'bg-blue-50 text-blue-700 border-blue-200'
                    }`}>
                      {key.type}
                    </span>
                  </td>
                  <td className="px-6 py-4.5 whitespace-nowrap text-sm text-gray-500">
                    <span className={`inline-flex items-center px-2.5 py-1 rounded-full text-xs font-semibold border ${
                      key.status === 'active' 
                        ? 'bg-green-50 text-green-700 border-green-200' 
                        : 'bg-red-50 text-red-700 border-red-200'
                    }`}>
                      {key.status}
                    </span>
                  </td>
                  <td className="px-6 py-4.5 whitespace-nowrap text-sm text-gray-500 font-medium">
                    {new Date(key.created_at).toLocaleDateString()}
                  </td>
                  <td className="px-6 py-4.5 whitespace-nowrap text-right text-sm font-medium">
                    {key.status === 'active' ? (
                      <button 
                        onClick={() => handleRevokeKey(key.id)}
                        className="text-xs font-semibold px-3 py-1.5 rounded-lg border border-red-200 text-red-600 hover:bg-red-50 transition-all duration-200 hover:shadow-xs cursor-pointer"
                      >
                        Revoke
                      </button>
                    ) : (
                      <span className="text-xs text-gray-400 font-semibold px-3 py-1.5 inline-block">Revoked</span>
                    )}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {showCreateModal && (
        <div className="fixed inset-0 bg-gray-900/40 backdrop-blur-xs flex items-center justify-center z-50 animate-fade-in">
          <div className="bg-white rounded-2xl shadow-xl max-w-md w-full mx-4 border border-gray-100 overflow-hidden">
            <div className="px-6 py-5 border-b border-gray-50 flex items-center justify-between">
              <h3 className="text-lg font-bold text-gray-900">Create API Key</h3>
              <button 
                onClick={() => setShowCreateModal(false)}
                className="text-gray-400 hover:text-gray-600 rounded-lg p-1 hover:bg-gray-55"
              >
                <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
            <form onSubmit={handleCreateKey}>
              <div className="p-6 space-y-4">
                <div>
                  <label className="block text-sm font-semibold text-gray-700 mb-1.5">Name</label>
                  <input
                    type="text"
                    required
                    className="block w-full px-3.5 py-2.5 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary-500 focus:border-transparent focus:outline-none transition-all duration-200 text-sm"
                    placeholder="Production API Key"
                    value={newName}
                    onChange={(e) => setNewName(e.target.value)}
                  />
                </div>
                <div>
                  <label className="block text-sm font-semibold text-gray-700 mb-1.5">Type</label>
                  <select 
                    className="block w-full px-3.5 py-2.5 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary-500 focus:border-transparent focus:outline-none transition-all duration-200 text-sm bg-white"
                    value={newType}
                    onChange={(e) => setNewType(e.target.value)}
                  >
                    <option value="test">Test</option>
                    <option value="live">Live</option>
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-semibold text-gray-700 mb-1.5">Scopes (Comma separated)</label>
                  <input
                    type="text"
                    required
                    className="block w-full px-3.5 py-2.5 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary-500 focus:border-transparent focus:outline-none transition-all duration-200 text-sm"
                    placeholder="tenant.read,tenant.write,users.read"
                    value={newScopes}
                    onChange={(e) => setNewScopes(e.target.value)}
                  />
                </div>
              </div>
              <div className="px-6 py-4.5 bg-gray-50 border-t border-gray-100 flex justify-end space-x-3">
                <button
                  type="button"
                  onClick={() => setShowCreateModal(false)}
                  className="px-4 py-2 border border-gray-200 rounded-xl text-sm font-semibold text-gray-600 bg-white hover:bg-gray-50 transition-all duration-200 cursor-pointer"
                >
                  Cancel
                </button>
                <button 
                  type="submit"
                  className="px-4 py-2 bg-primary-600 text-white rounded-xl text-sm font-semibold hover:bg-primary-700 transition-all duration-200 shadow-sm cursor-pointer"
                >
                  Create
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
