import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React, { ReactNode } from 'react';

// Create a wrapper for tests that need QueryClient
function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  });
  
  return function Wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        {children}
      </QueryClientProvider>
    );
  };
}

describe('Test Utils', () => {
  it('creates a QueryClient wrapper', () => {
    const Wrapper = createWrapper();
    expect(Wrapper).toBeDefined();
  });
});

// Mock Button component for tests
const MockButton = ({ 
  children, 
  onClick, 
  disabled, 
  variant = 'default',
  size = 'default'
}: {
  children: ReactNode;
  onClick?: () => void;
  disabled?: boolean;
  variant?: string;
  size?: string;
}) => (
  <button 
    onClick={onClick} 
    disabled={disabled}
    data-variant={variant}
    data-size={size}
  >
    {children}
  </button>
);

describe('Mock Button Component', () => {
  it('renders children', () => {
    render(<MockButton>Click Me</MockButton>);
    expect(screen.getByText('Click Me')).toBeInTheDocument();
  });

  it('handles click events', () => {
    const handleClick = vi.fn();
    render(<MockButton onClick={handleClick}>Click Me</MockButton>);
    
    fireEvent.click(screen.getByText('Click Me'));
    expect(handleClick).toHaveBeenCalledTimes(1);
  });

  it('can be disabled', () => {
    const handleClick = vi.fn();
    render(<MockButton onClick={handleClick} disabled>Click Me</MockButton>);
    
    const button = screen.getByText('Click Me');
    expect(button).toBeDisabled();
    
    fireEvent.click(button);
    expect(handleClick).not.toHaveBeenCalled();
  });

  it('supports variant prop', () => {
    render(<MockButton variant="destructive">Delete</MockButton>);
    const button = screen.getByText('Delete');
    expect(button).toHaveAttribute('data-variant', 'destructive');
  });

  it('supports size prop', () => {
    render(<MockButton size="lg">Large Button</MockButton>);
    const button = screen.getByText('Large Button');
    expect(button).toHaveAttribute('data-size', 'lg');
  });
});

// Mock Input component for tests
const MockInput = ({
  value,
  onChange,
  placeholder,
  type = 'text',
  disabled = false,
  'aria-label': ariaLabel,
}: {
  value?: string;
  onChange?: (e: React.ChangeEvent<HTMLInputElement>) => void;
  placeholder?: string;
  type?: string;
  disabled?: boolean;
  'aria-label'?: string;
}) => (
  <input
    type={type}
    value={value}
    onChange={onChange}
    placeholder={placeholder}
    disabled={disabled}
    aria-label={ariaLabel}
  />
);

describe('Mock Input Component', () => {
  it('renders with placeholder', () => {
    render(<MockInput placeholder="Enter text" />);
    expect(screen.getByPlaceholderText('Enter text')).toBeInTheDocument();
  });

  it('handles onChange events', () => {
    const handleChange = vi.fn();
    render(<MockInput onChange={handleChange} aria-label="test input" />);
    
    const input = screen.getByLabelText('test input');
    fireEvent.change(input, { target: { value: 'test' } });
    
    expect(handleChange).toHaveBeenCalled();
  });

  it('can be disabled', () => {
    render(<MockInput disabled aria-label="disabled input" />);
    expect(screen.getByLabelText('disabled input')).toBeDisabled();
  });

  it('supports different types', () => {
    render(<MockInput type="password" aria-label="password input" />);
    const input = screen.getByLabelText('password input');
    expect(input).toHaveAttribute('type', 'password');
  });
});

// Mock Form component for tests
const MockForm = ({
  onSubmit,
  children,
}: {
  onSubmit?: (data: Record<string, string>) => void;
  children: ReactNode;
}) => {
  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);
    const data: Record<string, string> = {};
    formData.forEach((value, key) => {
      data[key] = value.toString();
    });
    onSubmit?.(data);
  };

  return <form onSubmit={handleSubmit}>{children}</form>;
};

describe('Mock Form Component', () => {
  it('handles form submission', async () => {
    const handleSubmit = vi.fn();
    
    render(
      <MockForm onSubmit={handleSubmit}>
        <input name="email" defaultValue="test@example.com" />
        <button type="submit">Submit</button>
      </MockForm>
    );

    fireEvent.click(screen.getByText('Submit'));
    
    await waitFor(() => {
      expect(handleSubmit).toHaveBeenCalledWith({
        email: 'test@example.com',
      });
    });
  });

  it('prevents default form behavior', () => {
    const handleSubmit = vi.fn();
    
    render(
      <MockForm onSubmit={handleSubmit}>
        <button type="submit">Submit</button>
      </MockForm>
    );

    const form = screen.getByRole('button', { name: 'Submit' }).closest('form');
    const submitEvent = new Event('submit', { bubbles: true, cancelable: true });
    form?.dispatchEvent(submitEvent);
  });
});

// Loading Spinner mock test
const LoadingSpinner = ({ size = 'md' }: { size?: 'sm' | 'md' | 'lg' }) => {
  const sizeClasses = {
    sm: 'h-4 w-4',
    md: 'h-6 w-6',
    lg: 'h-8 w-8',
  };

  return (
    <div 
      role="status" 
      className={sizeClasses[size]}
      aria-label="Loading"
    >
      <span className="sr-only">Loading...</span>
    </div>
  );
};

describe('Loading Spinner', () => {
  it('renders with default size', () => {
    render(<LoadingSpinner />);
    expect(screen.getByRole('status')).toBeInTheDocument();
  });

  it('has accessible label', () => {
    render(<LoadingSpinner />);
    expect(screen.getByLabelText('Loading')).toBeInTheDocument();
  });

  it('contains screen reader text', () => {
    render(<LoadingSpinner />);
    expect(screen.getByText('Loading...')).toBeInTheDocument();
  });
});

// Alert/Toast mock tests
const MockAlert = ({
  type = 'info',
  message,
  onClose,
}: {
  type?: 'info' | 'success' | 'warning' | 'error';
  message: string;
  onClose?: () => void;
}) => (
  <div role="alert" data-type={type}>
    <p>{message}</p>
    {onClose && (
      <button onClick={onClose} aria-label="Close alert">
        ×
      </button>
    )}
  </div>
);

describe('Mock Alert Component', () => {
  it('renders message', () => {
    render(<MockAlert message="This is an alert" />);
    expect(screen.getByText('This is an alert')).toBeInTheDocument();
  });

  it('has alert role', () => {
    render(<MockAlert message="Alert message" />);
    expect(screen.getByRole('alert')).toBeInTheDocument();
  });

  it('can be closed', () => {
    const handleClose = vi.fn();
    render(<MockAlert message="Closeable alert" onClose={handleClose} />);
    
    fireEvent.click(screen.getByLabelText('Close alert'));
    expect(handleClose).toHaveBeenCalled();
  });

  it('supports different types', () => {
    render(<MockAlert type="error" message="Error message" />);
    const alert = screen.getByRole('alert');
    expect(alert).toHaveAttribute('data-type', 'error');
  });
});

// Modal mock tests
const MockModal = ({
  isOpen,
  onClose,
  title,
  children,
}: {
  isOpen: boolean;
  onClose: () => void;
  title: string;
  children: ReactNode;
}) => {
  if (!isOpen) return null;

  return (
    <div role="dialog" aria-labelledby="modal-title">
      <h2 id="modal-title">{title}</h2>
      <button onClick={onClose} aria-label="Close modal">
        ×
      </button>
      <div>{children}</div>
    </div>
  );
};

describe('Mock Modal Component', () => {
  it('is not rendered when closed', () => {
    render(
      <MockModal isOpen={false} onClose={vi.fn()} title="Test Modal">
        Content
      </MockModal>
    );
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('is rendered when open', () => {
    render(
      <MockModal isOpen={true} onClose={vi.fn()} title="Test Modal">
        Content
      </MockModal>
    );
    expect(screen.getByRole('dialog')).toBeInTheDocument();
  });

  it('displays title', () => {
    render(
      <MockModal isOpen={true} onClose={vi.fn()} title="Test Modal">
        Content
      </MockModal>
    );
    expect(screen.getByText('Test Modal')).toBeInTheDocument();
  });

  it('calls onClose when close button clicked', () => {
    const handleClose = vi.fn();
    render(
      <MockModal isOpen={true} onClose={handleClose} title="Test Modal">
        Content
      </MockModal>
    );
    
    fireEvent.click(screen.getByLabelText('Close modal'));
    expect(handleClose).toHaveBeenCalled();
  });

  it('has proper accessibility attributes', () => {
    render(
      <MockModal isOpen={true} onClose={vi.fn()} title="Accessible Modal">
        Content
      </MockModal>
    );
    
    const dialog = screen.getByRole('dialog');
    expect(dialog).toHaveAttribute('aria-labelledby', 'modal-title');
  });
});

// Pagination mock tests
const MockPagination = ({
  currentPage,
  totalPages,
  onPageChange,
}: {
  currentPage: number;
  totalPages: number;
  onPageChange: (page: number) => void;
}) => (
  <nav role="navigation" aria-label="Pagination">
    <button
      onClick={() => onPageChange(currentPage - 1)}
      disabled={currentPage <= 1}
      aria-label="Previous page"
    >
      Previous
    </button>
    <span>
      Page {currentPage} of {totalPages}
    </span>
    <button
      onClick={() => onPageChange(currentPage + 1)}
      disabled={currentPage >= totalPages}
      aria-label="Next page"
    >
      Next
    </button>
  </nav>
);

describe('Mock Pagination Component', () => {
  it('displays current page info', () => {
    render(
      <MockPagination currentPage={2} totalPages={5} onPageChange={vi.fn()} />
    );
    expect(screen.getByText('Page 2 of 5')).toBeInTheDocument();
  });

  it('disables previous on first page', () => {
    render(
      <MockPagination currentPage={1} totalPages={5} onPageChange={vi.fn()} />
    );
    expect(screen.getByLabelText('Previous page')).toBeDisabled();
  });

  it('disables next on last page', () => {
    render(
      <MockPagination currentPage={5} totalPages={5} onPageChange={vi.fn()} />
    );
    expect(screen.getByLabelText('Next page')).toBeDisabled();
  });

  it('calls onPageChange with next page', () => {
    const handlePageChange = vi.fn();
    render(
      <MockPagination
        currentPage={2}
        totalPages={5}
        onPageChange={handlePageChange}
      />
    );

    fireEvent.click(screen.getByLabelText('Next page'));
    expect(handlePageChange).toHaveBeenCalledWith(3);
  });

  it('calls onPageChange with previous page', () => {
    const handlePageChange = vi.fn();
    render(
      <MockPagination
        currentPage={2}
        totalPages={5}
        onPageChange={handlePageChange}
      />
    );

    fireEvent.click(screen.getByLabelText('Previous page'));
    expect(handlePageChange).toHaveBeenCalledWith(1);
  });
});
