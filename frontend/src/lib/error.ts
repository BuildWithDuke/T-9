export interface ErrorDetails {
  type: string;
  message: string;
  code?: number;
  details?: string;
  timestamp: number;
}

export class AppError extends Error {
  public readonly type: string;
  public readonly code?: number;
  public readonly details?: string;
  public readonly timestamp: number;

  constructor(
    type: string,
    message: string,
    code?: number,
    details?: string
  ) {
    super(message);
    this.name = 'AppError';
    this.type = type;
    this.code = code;
    this.details = details;
    this.timestamp = Date.now();
  }

  toJSON(): ErrorDetails {
    return {
      type: this.type,
      message: this.message,
      code: this.code,
      details: this.details,
      timestamp: this.timestamp,
    };
  }

  getUserMessage(): string {
    switch (this.type) {
      case 'NETWORK_ERROR':
        return 'Network connection error. Please check your internet connection.';
      case 'SERVER_ERROR':
        return 'Server error occurred. Please try again later.';
      case 'VALIDATION_ERROR':
        return this.message;
      case 'GAME_ERROR':
        return this.message;
      case 'WEBSOCKET_ERROR':
        return 'Connection to game server lost. Please refresh the page.';
      default:
        return 'An unexpected error occurred. Please try again.';
    }
  }
}

// Error types
export const ErrorTypes = {
  NETWORK_ERROR: 'NETWORK_ERROR',
  SERVER_ERROR: 'SERVER_ERROR',
  VALIDATION_ERROR: 'VALIDATION_ERROR',
  GAME_ERROR: 'GAME_ERROR',
  WEBSOCKET_ERROR: 'WEBSOCKET_ERROR',
  AUTH_ERROR: 'AUTH_ERROR',
} as const;

// Error factory functions
export const ErrorFactory = {
  network: (message: string, details?: string) =>
    new AppError(ErrorTypes.NETWORK_ERROR, message, undefined, details),
  
  server: (message: string, code?: number, details?: string) =>
    new AppError(ErrorTypes.SERVER_ERROR, message, code, details),
  
  validation: (message: string, details?: string) =>
    new AppError(ErrorTypes.VALIDATION_ERROR, message, undefined, details),
  
  game: (message: string, details?: string) =>
    new AppError(ErrorTypes.GAME_ERROR, message, undefined, details),
  
  websocket: (message: string, details?: string) =>
    new AppError(ErrorTypes.WEBSOCKET_ERROR, message, undefined, details),
  
  auth: (message: string, details?: string) =>
    new AppError(ErrorTypes.AUTH_ERROR, message, undefined, details),
};

// Error handler utility
export class ErrorHandler {
  private static errors: ErrorDetails[] = [];
  private static maxErrors = 50;

  static handleError(error: Error | AppError | string): AppError {
    let appError: AppError;

    if (error instanceof AppError) {
      appError = error;
    } else if (error instanceof Error) {
      appError = ErrorFactory.server(error.message, undefined, error.stack);
    } else if (typeof error === 'string') {
      appError = ErrorFactory.validation(error);
    } else {
      appError = ErrorFactory.server('Unknown error occurred');
    }

    // Add to error log
    this.errors.push(appError.toJSON());
    
    // Keep only recent errors
    if (this.errors.length > this.maxErrors) {
      this.errors = this.errors.slice(-this.maxErrors);
    }

    // Log to console in development
    if (import.meta.env.DEV) {
      console.error('App Error:', appError);
    }

    return appError;
  }

  static getErrorLog(): ErrorDetails[] {
    return [...this.errors];
  }

  static clearErrorLog(): void {
    this.errors = [];
  }

  static isNetworkError(error: Error): boolean {
    return error.message.includes('Network') || 
           error.message.includes('fetch') ||
           error.message.includes('ECONNREFUSED');
  }

  static isServerError(error: Error): boolean {
    return error.message.includes('500') ||
           error.message.includes('Internal Server Error');
  }
}

// API error handler
export class APIError extends AppError {
  constructor(
    public readonly statusCode: number,
    message: string,
    details?: string
  ) {
    super('API_ERROR', message, statusCode, details);
    this.name = 'APIError';
  }

  static async fromResponse(response: Response): Promise<APIError> {
    let message = `HTTP ${response.status}`;
    let details = '';

    try {
      const data = await response.json();
      if (data.error) {
        message = data.error.message || data.error;
        details = data.error.details || '';
      }
    } catch {
      // Use default message if can't parse JSON
    }

    return new APIError(response.status, message, details);
  }
}

// Safe async wrapper
export async function safeAsync<T>(
  fn: () => Promise<T>,
  errorHandler?: (error: AppError) => void
): Promise<T | null> {
  try {
    return await fn();
  } catch (error) {
    const appError = ErrorHandler.handleError(error);
    if (errorHandler) {
      errorHandler(appError);
    }
    return null;
  }
}

// Safe fetch wrapper
export async function safeFetch(
  url: string,
  options?: RequestInit
): Promise<Response> {
  try {
    const response = await fetch(url, options);
    
    if (!response.ok) {
      throw await APIError.fromResponse(response);
    }
    
    return response;
  } catch (error) {
    if (error instanceof APIError) {
      throw error;
    }
    throw ErrorHandler.handleError(error);
  }
}