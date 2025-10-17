'use client';

import { evaluatePasswordStrength } from '@/lib/password';

interface PasswordStrengthMeterProps {
  password: string;
}

const strengthColors = ['bg-red-500', 'bg-orange-500', 'bg-yellow-500', 'bg-green-500'];

export function PasswordStrengthMeter({ password }: PasswordStrengthMeterProps) {
  if (!password) {
    return null;
  }
  const strength = evaluatePasswordStrength(password);

  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2">
        <div className="flex flex-1 h-2 rounded-full bg-gray-200 overflow-hidden">
          {[0, 1, 2, 3].map((index) => (
            <span
              key={index}
              className={`flex-1 transition-colors duration-300 ${
                strength.score > index ? strengthColors[Math.min(strength.score - 1, strengthColors.length - 1)] : 'bg-gray-200'
              }`}
            />
          ))}
        </div>
        <span className="text-xs font-medium text-gray-600 w-20 text-right">{strength.label}</span>
      </div>
      {!strength.isAcceptable && strength.suggestions.length > 0 && (
        <p className="text-xs text-gray-500">
          Suggestions: {strength.suggestions.join(', ')}
        </p>
      )}
    </div>
  );
}
