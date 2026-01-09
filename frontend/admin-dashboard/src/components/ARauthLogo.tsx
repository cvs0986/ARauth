/**
 * ARauth Logo Component
 * A unique, modern logo for the ARauth Identity Platform
 */

export function ARauthLogo({ className = '', size = 'md' }: { className?: string; size?: 'sm' | 'md' | 'lg' }) {
  const sizeClasses = {
    sm: 'w-8 h-8',
    md: 'w-12 h-12',
    lg: 'w-16 h-16',
  };

  return (
    <div className={`flex items-center gap-2 ${className}`}>
      <div className={`${sizeClasses[size]} relative`}>
        {/* Main logo shape - Shield with key */}
        <svg
          viewBox="0 0 64 64"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
          className="w-full h-full"
        >
          {/* Shield background with gradient */}
          <defs>
            <linearGradient id="shieldGradient" x1="0%" y1="0%" x2="100%" y2="100%">
              <stop offset="0%" stopColor="#3b82f6" />
              <stop offset="100%" stopColor="#2563eb" />
            </linearGradient>
            <linearGradient id="keyGradient" x1="0%" y1="0%" x2="100%" y2="100%">
              <stop offset="0%" stopColor="#fbbf24" />
              <stop offset="100%" stopColor="#f59e0b" />
            </linearGradient>
          </defs>
          
          {/* Shield shape */}
          <path
            d="M32 4 L12 12 L12 28 C12 40 20 50 32 56 C44 50 52 40 52 28 L52 12 Z"
            fill="url(#shieldGradient)"
            stroke="#1e40af"
            strokeWidth="2"
          />
          
          {/* Key shape inside shield */}
          <path
            d="M32 24 C28 24 25 27 25 31 C25 35 28 38 32 38 C36 38 39 35 39 31 C39 27 36 24 32 24 Z M32 28 L36 28 L36 30 L34 30 L34 34 L36 34 L36 36 L32 36 L32 28 Z"
            fill="url(#keyGradient)"
          />
          
          {/* Decorative circle */}
          <circle cx="32" cy="20" r="3" fill="#ffffff" opacity="0.8" />
        </svg>
      </div>
      <div className="flex flex-col">
        <span className={`font-bold text-primary-600 ${size === 'sm' ? 'text-lg' : size === 'md' ? 'text-xl' : 'text-2xl'}`}>
          ARauth
        </span>
        <span className={`text-secondary-500 ${size === 'sm' ? 'text-xs' : size === 'md' ? 'text-sm' : 'text-base'}`}>
          Identity Platform
        </span>
      </div>
    </div>
  );
}


