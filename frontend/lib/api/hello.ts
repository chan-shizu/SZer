type GetHelloResponse = {
  message: string;
};

export async function getHello(): Promise<GetHelloResponse> {
  const url = process.env.API_BASE_URL || "http://localhost:8080";
  console.log("Fetching from URL:", url);

  const res = await fetch(url, {
    method: "GET",
    headers: { Accept: "application/json" },
  });

  if (!res.ok) {
    throw new Error(`Request failed with status ${res.status}`);
  }

  return res.json();
}

export default getHello;
