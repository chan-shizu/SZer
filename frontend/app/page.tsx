export default async function Home() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-subtle font-sans dark:bg-black">
      <main className="flex min-h-screen w-full max-w-3xl flex-col items-center justify-between py-32 px-16 bg-background dark:bg-black sm:items-start">
        <h1 className="text-5xl font-extrabold text-foreground dark:text-zinc-100 sm:text-6xl">
          Welcome to <span className="text-blue-600">SZer</span>
        </h1>
      </main>
    </div>
  );
}
