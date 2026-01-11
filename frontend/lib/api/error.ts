export type ApiErrorResponse = {
  code: string;
  message: string;
  details?: unknown;
  requestId?: string;
};

export class ApiError extends Error {
  status: number;
  code: string;
  url: string;
  details?: unknown;
  requestId?: string;

  constructor(args: {
    status: number;
    code: string;
    message: string;
    url: string;
    details?: unknown;
    requestId?: string;
  }) {
    super(args.message);
    this.name = "ApiError";
    this.status = args.status;
    this.code = args.code;
    this.url = args.url;
    this.details = args.details;
    this.requestId = args.requestId;
  }
}
