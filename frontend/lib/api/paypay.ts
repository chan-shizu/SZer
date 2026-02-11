export type PayPayCheckoutResponse = {
  merchant_payment_id: string;
  url: string;
  deeplink: string;
};

export type PayPayConfirmResponse = {
  status: string;
  granted: boolean;
  program_id: number;
};

export async function createPayPayCheckout(programId: number): Promise<PayPayCheckoutResponse> {
  const res = await fetch("/api/me/paypay/checkout", {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ program_id: programId }),
  });

  if (!res.ok) {
    const text = await res.text().catch(() => "");
    throw new Error(text || `Request failed with status ${res.status}`);
  }

  return (await res.json()) as PayPayCheckoutResponse;
}

export async function confirmPayPayPayment(merchantPaymentId: string): Promise<PayPayConfirmResponse> {
  const res = await fetch(`/api/me/paypay/payments/${encodeURIComponent(merchantPaymentId)}`, {
    method: "GET",
    credentials: "include",
  });

  if (!res.ok) {
    const text = await res.text().catch(() => "");
    throw new Error(text || `Request failed with status ${res.status}`);
  }

  return (await res.json()) as PayPayConfirmResponse;
}
