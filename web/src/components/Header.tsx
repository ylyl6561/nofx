import { useLanguage } from '../contexts/LanguageContext'
import { t } from '../i18n/translations'
import { useEffect, useState } from 'react'

interface HeaderProps {
  simple?: boolean // For login/register pages
}

export function Header({ simple = false }: HeaderProps) {
  const { language, setLanguage } = useLanguage()
  const [totalTokens, setTotalTokens] = useState<number>(0)
  const [loading, setLoading] = useState(false)

  // 获取用户总 token 使用量
  useEffect(() => {
    console.log('[Header] useEffect triggered, simple:', simple)
    
    if (simple) {
      console.log('[Header] Simple mode, skipping token fetch')
      return // 简单模式下不显示 token
    }

    const fetchTokenUsage = async () => {
      try {
        setLoading(true)
        const token = localStorage.getItem('token')
        console.log('[Header] Token from localStorage:', token ? 'exists' : 'missing')
        
        if (!token) {
          console.warn('[Header] No token found, skipping API call')
          return
        }

        console.log('[Header] Fetching token usage from /api/users/token-usage')
        const response = await fetch('/api/users/token-usage', {
          headers: {
            'Authorization': `Bearer ${token}`
          }
        })

        console.log('[Header] API response status:', response.status)
        
        if (response.ok) {
          const data = await response.json()
          console.log('[Header] Token usage response:', data)
          setTotalTokens(data.total_tokens || 0)
        } else {
          console.error('[Header] Token usage API error:', response.status, response.statusText)
          const errorText = await response.text()
          console.error('[Header] Error response:', errorText)
        }
      } catch (error) {
        console.error('[Header] 获取 token 使用量失败:', error)
      } finally {
        setLoading(false)
      }
    }

    fetchTokenUsage()
    
    // 每分钟刷新一次
    const interval = setInterval(fetchTokenUsage, 60000)
    return () => clearInterval(interval)
  }, [simple])

  return (
    <header className="glass sticky top-0 z-50 backdrop-blur-xl">
      <div className="max-w-[1920px] mx-auto px-6 py-4">
        <div className="flex items-center justify-between">
          {/* Left - Logo and Title */}
          <div className="flex items-center gap-3">
            <div className="flex items-center justify-center">
              <img src="/icons/nofx.svg" alt="NoFx Logo" className="w-8 h-8" />
            </div>
            <div>
              <h1 className="text-xl font-bold" style={{ color: '#EAECEF' }}>
                {t('appTitle', language)}
              </h1>
              {!simple && (
                <p className="text-xs mono" style={{ color: '#848E9C' }}>
                  {t('subtitle', language)}
                </p>
              )}
            </div>
          </div>

          {/* Right - Token Usage & Language Toggle */}
          <div className="flex items-center gap-3">
            {/* Token Usage Display */}
            {!simple && (
              <div 
                className="flex items-center gap-2 px-4 py-2 rounded-lg"
                style={{ background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)' }}
              >
                <span className="text-lg">🎯</span>
                <div className="flex flex-col">
                  <span className="text-xs" style={{ color: '#E0E0E0' }}>
                    {language === 'zh' ? 'Token 使用量' : 'Token Usage'}
                  </span>
                  <span className="text-sm font-bold" style={{ color: '#FFFFFF' }}>
                    {loading ? '...' : totalTokens.toLocaleString()}
                  </span>
                </div>
              </div>
            )}

            {/* Language Toggle */}
            <div
              className="flex gap-1 rounded p-1"
              style={{ background: '#1E2329' }}
            >
              <button
              onClick={() => setLanguage('zh')}
              className="px-3 py-1.5 rounded text-xs font-semibold transition-all"
              style={
                language === 'zh'
                  ? { background: '#F0B90B', color: '#000' }
                  : { background: 'transparent', color: '#848E9C' }
              }
            >
              中文
            </button>
              <button
                onClick={() => setLanguage('en')}
                className="px-3 py-1.5 rounded text-xs font-semibold transition-all"
                style={
                  language === 'en'
                    ? { background: '#F0B90B', color: '#000' }
                    : { background: 'transparent', color: '#848E9C' }
                }
              >
                EN
              </button>
            </div>
          </div>
        </div>
      </div>
    </header>
  )
}
