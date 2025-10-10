export interface PasswordStrength {
  score: number;
  label: 'Very weak' | 'Weak' | 'Fair' | 'Strong' | 'Very strong';
  suggestions: string[];
  isAcceptable: boolean;
}

const requirements = [
  { test: /[a-z]/, message: 'Add a lowercase letter' },
  { test: /[A-Z]/, message: 'Add an uppercase letter' },
  { test: /[0-9]/, message: 'Add a number' },
  { test: /[^A-Za-z0-9]/, message: 'Add a special character' },
];

const commonPasswords = new Set([
  'password',
  '123456',
  '123456789',
  'qwerty',
  'abc123',
  'letmein',
  'welcome',
  'admin',
]);

export function evaluatePasswordStrength(password: string): PasswordStrength {
  const trimmed = password.trim();
  if (!trimmed) {
    return {
      score: 0,
      label: 'Very weak',
      suggestions: ['Password cannot be empty'],
      isAcceptable: false,
    };
  }

  const suggestions: string[] = [];
  let score = 0;

  if (commonPasswords.has(trimmed.toLowerCase())) {
    suggestions.push('Avoid using common passwords');
  }

  const length = trimmed.length;
  if (length >= 16) {
    score += 2;
  } else if (length >= 12) {
    score += 1.5;
  } else if (length >= 10) {
    score += 1;
  } else if (length >= 8) {
    score += 0.5;
    suggestions.push('Use at least 12 characters for better strength');
  } else {
    suggestions.push('Use at least 10 characters');
  }

  requirements.forEach((requirement) => {
    if (requirement.test.test(trimmed)) {
      score += 1;
    } else {
      suggestions.push(requirement.message);
    }
  });

  if (/([a-zA-Z0-9])\1{2,}/.test(trimmed)) {
    suggestions.push('Avoid repeated sequences');
    score -= 0.5;
  }

  if (/^[A-Za-z]+$/.test(trimmed) || /^[0-9]+$/.test(trimmed)) {
    suggestions.push('Mix letters, numbers, and symbols');
    score -= 0.5;
  }

  if (score < 0) {
    score = 0;
  }

  // Normalize to range 0-4
  const normalizedScore = Math.max(0, Math.min(4, Math.round(score)));

  let label: PasswordStrength['label'] = 'Very weak';
  switch (normalizedScore) {
    case 0:
      label = 'Very weak';
      break;
    case 1:
      label = 'Weak';
      break;
    case 2:
      label = 'Fair';
      break;
    case 3:
      label = 'Strong';
      break;
    case 4:
      label = 'Very strong';
      break;
    default:
      label = 'Very weak';
  }

  const isAcceptable = normalizedScore >= 3;

  return {
    score: normalizedScore,
    label,
    suggestions,
    isAcceptable,
  };
}
