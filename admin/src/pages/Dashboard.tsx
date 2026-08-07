import { Link } from 'react-router-dom';

export default function Dashboard() {
  // Live counts matched with the other pages' initial seed data
  const totalUsers = 5;
  const totalTenants = 3;
  const totalApiKeys = 2;

  const recentActivities = [
    { id: 'act_1', action: 'user.created', user: 'admin@saaskit.dev', time: '10 mins ago', status: 'Success' },
    { id: 'act_2', action: 'tenant.created', user: 'admin@saaskit.dev', time: '45 mins ago', status: 'Success' },
    { id: 'act_3', action: 'api_key.created', user: 'developer@saaskit.dev', time: '2 hours ago', status: 'Success' }
  ];

  return (
    <div className="animate-fade-in space-y-8">
      <div>
        <h1 className="text-3xl font-bold text-gray-900 tracking-tight">Dashboard</h1>
        <p className="mt-2 text-gray-600">
          Welcome to the SaaSKit Admin Console. Monitor and manage your platform instance.
        </p>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="bg-white rounded-2xl shadow-xs border border-gray-100 p-6 hover:shadow-sm transition-all duration-250">
          <div className="flex items-center">
            <div className="flex-shrink-0">
              <div className="w-12 h-12 rounded-xl bg-primary-50 border border-primary-100 flex items-center justify-center">
                <svg className="w-6 h-6 text-primary-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                </svg>
              </div>
            </div>
            <div className="ml-5 w-0 flex-1">
              <dl>
                <dt className="text-sm font-semibold text-gray-500 uppercase tracking-wider">
                  Total Users
                </dt>
                <dd className="text-3xl font-bold text-gray-900 mt-1">{totalUsers}</dd>
              </dl>
            </div>
          </div>
        </div>

        <div className="bg-white rounded-2xl shadow-xs border border-gray-100 p-6 hover:shadow-sm transition-all duration-250">
          <div className="flex items-center">
            <div className="flex-shrink-0">
              <div className="w-12 h-12 rounded-xl bg-green-50 border border-green-100 flex items-center justify-center">
                <svg className="w-6 h-6 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
                </svg>
              </div>
            </div>
            <div className="ml-5 w-0 flex-1">
              <dl>
                <dt className="text-sm font-semibold text-gray-500 uppercase tracking-wider">
                  Active Tenants
                </dt>
                <dd className="text-3xl font-bold text-gray-900 mt-1">{totalTenants}</dd>
              </dl>
            </div>
          </div>
        </div>

        <div className="bg-white rounded-2xl shadow-xs border border-gray-100 p-6 hover:shadow-sm transition-all duration-250">
          <div className="flex items-center">
            <div className="flex-shrink-0">
              <div className="w-12 h-12 rounded-xl bg-amber-50 border border-amber-100 flex items-center justify-center">
                <svg className="w-6 h-6 text-amber-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
                </svg>
              </div>
            </div>
            <div className="ml-5 w-0 flex-1">
              <dl>
                <dt className="text-sm font-semibold text-gray-500 uppercase tracking-wider">
                  Active API Keys
                </dt>
                <dd className="text-3xl font-bold text-gray-900 mt-1">{totalApiKeys}</dd>
              </dl>
            </div>
          </div>
        </div>
      </div>

      {/* Main Content Area */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Quick Actions & System Info */}
        <div className="lg:col-span-2 space-y-6">
          <div className="bg-white rounded-2xl shadow-xs border border-gray-100 overflow-hidden">
            <div className="px-6 py-5 border-b border-gray-50 bg-gray-50/50">
              <h3 className="text-lg font-bold text-gray-900">
                Quick Actions
              </h3>
              <p className="mt-1 text-sm text-gray-500">
                Fast-track common administrative tasks
              </p>
            </div>
            <div className="p-6">
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <Link 
                  to="/users" 
                  className="flex items-center p-4 border border-gray-100 rounded-xl hover:border-primary-100 hover:bg-primary-50/30 transition-all duration-200"
                >
                  <div className="w-10 h-10 rounded-lg bg-primary-50 flex items-center justify-center text-primary-600 mr-4">
                    <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z" />
                    </svg>
                  </div>
                  <div>
                    <h4 className="font-semibold text-gray-900 text-sm">Add New User</h4>
                    <p className="text-xs text-gray-500 mt-0.5">Register a new client profile</p>
                  </div>
                </Link>

                <Link 
                  to="/tenants" 
                  className="flex items-center p-4 border border-gray-100 rounded-xl hover:border-green-100 hover:bg-green-50/30 transition-all duration-200"
                >
                  <div className="w-10 h-10 rounded-lg bg-green-50 flex items-center justify-center text-green-600 mr-4">
                    <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                    </svg>
                  </div>
                  <div>
                    <h4 className="font-semibold text-gray-900 text-sm">Create Tenant</h4>
                    <p className="text-xs text-gray-500 mt-0.5">Initialize a new organization workspace</p>
                  </div>
                </Link>

                <Link 
                  to="/api-keys" 
                  className="flex items-center p-4 border border-gray-100 rounded-xl hover:border-amber-100 hover:bg-amber-50/30 transition-all duration-200"
                >
                  <div className="w-10 h-10 rounded-lg bg-amber-50 flex items-center justify-center text-amber-600 mr-4">
                    <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
                    </svg>
                  </div>
                  <div>
                    <h4 className="font-semibold text-gray-900 text-sm">Issue API Key</h4>
                    <p className="text-xs text-gray-500 mt-0.5">Generate credentials for developers</p>
                  </div>
                </Link>

                <Link 
                  to="/audit" 
                  className="flex items-center p-4 border border-gray-100 rounded-xl hover:border-purple-100 hover:bg-purple-50/30 transition-all duration-200"
                >
                  <div className="w-10 h-10 rounded-lg bg-purple-50 flex items-center justify-center text-purple-600 mr-4">
                    <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                    </svg>
                  </div>
                  <div>
                    <h4 className="font-semibold text-gray-900 text-sm">Audit Trail</h4>
                    <p className="text-xs text-gray-500 mt-0.5">Browse system activity logs</p>
                  </div>
                </Link>
              </div>
            </div>
          </div>

          {/* Recent Activity Table */}
          <div className="bg-white rounded-2xl shadow-xs border border-gray-100 overflow-hidden">
            <div className="px-6 py-5 border-b border-gray-50 flex items-center justify-between bg-gray-50/50">
              <div>
                <h3 className="text-lg font-bold text-gray-900">Recent Activity</h3>
                <p className="text-xs text-gray-500 mt-0.5">Real-time actions logs</p>
              </div>
              <Link to="/audit" className="text-xs font-semibold text-primary-600 hover:text-primary-700">
                View all logs →
              </Link>
            </div>
            <div className="divide-y divide-gray-50">
              {recentActivities.map((act) => (
                <div key={act.id} className="p-4 px-6 flex items-center justify-between hover:bg-gray-50/40 transition-colors duration-150">
                  <div className="flex items-center space-x-3">
                    <div className="w-2.5 h-2.5 rounded-full bg-primary-500 animate-pulse"></div>
                    <div>
                      <div className="text-sm font-semibold text-gray-800">{act.action}</div>
                      <div className="text-xs text-gray-400 mt-0.5">by {act.user}</div>
                    </div>
                  </div>
                  <div className="flex items-center space-x-3">
                    <span className="text-xs text-gray-400 font-medium">{act.time}</span>
                    <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold bg-green-50 text-green-700 border border-green-200">
                      {act.status}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Sidebar Status Info */}
        <div className="space-y-6">
          <div className="bg-white rounded-2xl shadow-xs border border-gray-100 p-6 space-y-6">
            <h3 className="text-lg font-bold text-gray-900">System Status</h3>
            <div className="space-y-4">
              <div className="flex justify-between items-center pb-3 border-b border-gray-50">
                <span className="text-sm text-gray-500 font-medium">Gateway API</span>
                <span className="inline-flex items-center text-xs font-semibold text-green-700 bg-green-50 px-2 py-0.5 rounded-md border border-green-150">
                  Online
                </span>
              </div>
              <div className="flex justify-between items-center pb-3 border-b border-gray-50">
                <span className="text-sm text-gray-500 font-medium">Database Node</span>
                <span className="inline-flex items-center text-xs font-semibold text-green-700 bg-green-50 px-2 py-0.5 rounded-md border border-green-150">
                  Healthy
                </span>
              </div>
              <div className="flex justify-between items-center pb-3 border-b border-gray-50">
                <span className="text-sm text-gray-500 font-medium">OIDC Issuer</span>
                <span className="inline-flex items-center text-xs font-semibold text-green-700 bg-green-50 px-2 py-0.5 rounded-md border border-green-150">
                  Ready
                </span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-gray-500 font-medium">SDK Version</span>
                <span className="text-sm text-gray-700 font-semibold font-mono">v0.1.0</span>
              </div>
            </div>
          </div>

          <div className="bg-gradient-to-br from-primary-600 to-primary-800 rounded-2xl shadow-sm p-6 text-white relative overflow-hidden">
            <div className="absolute right-0 bottom-0 translate-x-4 translate-y-4 opacity-10">
              <svg className="w-40 h-40" fill="currentColor" viewBox="0 0 24 24">
                <path d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
              </svg>
            </div>
            <h4 className="font-bold text-lg mb-2">Secure by Default</h4>
            <p className="text-xs text-primary-100 leading-relaxed">
              SaaSKit Admin automatically enforces enterprise encryption and granular authorization scoping rules.
            </p>
            <div className="mt-4">
              <a href="https://github.com/medaminerjb/saas-kit" target="_blank" rel="noreferrer" className="inline-block text-xs font-bold bg-white text-primary-700 px-3.5 py-2 rounded-xl shadow-xs hover:bg-primary-50 transition-colors duration-200">
                View Documentation
              </a>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
