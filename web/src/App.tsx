import { useEffect, useMemo, useState } from 'react'

type Game = {
  id: string
  name: string
  studio: string
  category: 'Strategy' | 'Puzzle' | 'Party' | 'Arcade' | 'Board' | 'Adventure'
  modes: Array<'Solo' | 'Online Multiplayer' | 'Offline Multiplayer' | 'Bot Match'>
  description: string
  cover: string
  rating: number
  players: string
  releaseYear: number
  price: number
  installed: boolean
  tags: string[]
}

const games: Game[] = [
  {
    id: 'mod-grid',
    name: 'Modifier Grid',
    studio: 'Northline Labs',
    category: 'Strategy',
    modes: ['Solo', 'Online Multiplayer', 'Bot Match'],
    description: 'Competitive grid tactics with modifier cards, sudden twists, and ranked ladders.',
    cover:
      'https://images.unsplash.com/photo-1511512578047-dfb367046420?auto=format&fit=crop&w=1400&q=80',
    rating: 4.8,
    players: '42k online',
    releaseYear: 2026,
    price: 0,
    installed: true,
    tags: ['ranked', 'modifiers', 'leaderboard', 'pvp'],
  },
  {
    id: 'rift-racers',
    name: 'Rift Racers',
    studio: 'Quiet Orbit',
    category: 'Arcade',
    modes: ['Online Multiplayer', 'Offline Multiplayer'],
    description: 'Fast lane duels in short rounds with drift boosts and spectator highlights.',
    cover:
      'https://images.unsplash.com/photo-1579373903781-fd5c0c30c4cd?auto=format&fit=crop&w=1400&q=80',
    rating: 4.6,
    players: '18k online',
    releaseYear: 2025,
    price: 5,
    installed: false,
    tags: ['racing', 'party', 'short-session'],
  },
  {
    id: 'tiny-kingdoms',
    name: 'Tiny Kingdoms',
    studio: 'Clay Beacon',
    category: 'Strategy',
    modes: ['Solo', 'Bot Match'],
    description: 'Asymmetric kingdom growth with weather events and compact macro decisions.',
    cover:
      'https://images.unsplash.com/photo-1542751371-adc38448a05e?auto=format&fit=crop&w=1400&q=80',
    rating: 4.7,
    players: '11k online',
    releaseYear: 2026,
    price: 8,
    installed: false,
    tags: ['city-building', 'simulation', 'deep'],
  },
  {
    id: 'echo-tiles',
    name: 'Echo Tiles',
    studio: 'Moss Arcade',
    category: 'Puzzle',
    modes: ['Solo', 'Online Multiplayer'],
    description: 'Rhythm-driven puzzle chains where each move reshapes future tile waves.',
    cover:
      'https://images.unsplash.com/photo-1550745165-9bc0b252726f?auto=format&fit=crop&w=1400&q=80',
    rating: 4.5,
    players: '9k online',
    releaseYear: 2024,
    price: 0,
    installed: true,
    tags: ['relaxing', 'brain', 'co-op'],
  },
  {
    id: 'couch-club',
    name: 'Couch Club',
    studio: 'Gather House',
    category: 'Party',
    modes: ['Offline Multiplayer', 'Online Multiplayer'],
    description: 'Mini-games for local nights with drop-in rooms and one-tap rematches.',
    cover:
      'https://images.unsplash.com/photo-1511882150382-421056c89033?auto=format&fit=crop&w=1400&q=80',
    rating: 4.4,
    players: '24k online',
    releaseYear: 2023,
    price: 0,
    installed: false,
    tags: ['party', 'friends', 'casual'],
  },
  {
    id: 'canvas-quest',
    name: 'Canvas Quest',
    studio: 'Signal Pine',
    category: 'Adventure',
    modes: ['Solo', 'Online Multiplayer'],
    description: 'Explore hand-painted worlds, decode ruins, and build cooperative camps.',
    cover:
      'https://images.unsplash.com/photo-1493711662062-fa541adb3fc8?auto=format&fit=crop&w=1400&q=80',
    rating: 4.9,
    players: '31k online',
    releaseYear: 2026,
    price: 12,
    installed: false,
    tags: ['story', 'co-op', 'exploration'],
  },
  {
    id: 'tableline',
    name: 'Tableline Duel',
    studio: 'Brass Root',
    category: 'Board',
    modes: ['Online Multiplayer', 'Bot Match'],
    description: 'Classic board mechanics reimagined with adaptive rule packs and drafts.',
    cover:
      'https://images.unsplash.com/photo-1528819622765-d6bcf132f793?auto=format&fit=crop&w=1400&q=80',
    rating: 4.3,
    players: '6k online',
    releaseYear: 2022,
    price: 0,
    installed: true,
    tags: ['board', 'turn-based', 'ranked'],
  },
]

const categories = ['All', 'Strategy', 'Puzzle', 'Party', 'Arcade', 'Board', 'Adventure'] as const
const modes = ['All', 'Solo', 'Online Multiplayer', 'Offline Multiplayer', 'Bot Match'] as const
const sortOptions = ['Trending', 'Rating', 'Newest', 'A-Z'] as const

function App() {
  const [query, setQuery] = useState('')
  const [category, setCategory] = useState<(typeof categories)[number]>('All')
  const [mode, setMode] = useState<(typeof modes)[number]>('All')
  const [freeOnly, setFreeOnly] = useState(false)
  const [installedOnly, setInstalledOnly] = useState(false)
  const [sortBy, setSortBy] = useState<(typeof sortOptions)[number]>('Trending')
  const [selectedGame, setSelectedGame] = useState<Game | null>(games[0])

  useEffect(() => {
    const listener = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'k') {
        event.preventDefault()
        const input = document.getElementById('store-search') as HTMLInputElement | null
        input?.focus()
      }
    }
    window.addEventListener('keydown', listener)
    return () => window.removeEventListener('keydown', listener)
  }, [])

  const filteredGames = useMemo(() => {
    const normalized = query.trim().toLowerCase()

    const result = games.filter((game) => {
      const matchQuery =
        normalized.length === 0 ||
        [game.name, game.studio, game.description, ...game.tags].join(' ').toLowerCase().includes(normalized)
      const matchCategory = category === 'All' || game.category === category
      const matchMode = mode === 'All' || game.modes.includes(mode)
      const matchPrice = !freeOnly || game.price === 0
      const matchInstalled = !installedOnly || game.installed

      return matchQuery && matchCategory && matchMode && matchPrice && matchInstalled
    })

    result.sort((a, b) => {
      if (sortBy === 'Rating') return b.rating - a.rating
      if (sortBy === 'Newest') return b.releaseYear - a.releaseYear
      if (sortBy === 'A-Z') return a.name.localeCompare(b.name)
      return b.players.localeCompare(a.players)
    })

    return result
  }, [category, freeOnly, installedOnly, mode, query, sortBy])

  const featured = filteredGames[0] ?? games[0]

  return (
    <div className="min-h-screen bg-stone-100 text-stone-900">
      <div className="mx-auto max-w-7xl px-4 pb-16 pt-8 sm:px-6 lg:px-8">
        <header className="mb-8 rounded-3xl border border-stone-200 bg-gradient-to-br from-stone-100 via-amber-50 to-orange-100 p-5 shadow-sm sm:p-7">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div>
              <p className="font-sans text-xs uppercase tracking-[0.24em] text-stone-600">HubGame Store</p>
              <h1 className="font-display mt-2 text-3xl leading-tight sm:text-4xl">Cozy collection for quick play and deep sessions</h1>
            </div>
            <p className="rounded-full border border-stone-300 bg-white/80 px-3 py-1 text-xs text-stone-600">{games.length} curated games</p>
          </div>

          <div className="mt-6 grid gap-5 lg:grid-cols-[1.5fr_1fr]">
            <div className="relative overflow-hidden rounded-2xl border border-stone-300 bg-stone-900 text-white">
              <img src={featured.cover} alt={featured.name} className="h-72 w-full object-cover opacity-60" />
              <div className="absolute inset-0 bg-gradient-to-t from-black/80 via-black/20 to-transparent" />
              <div className="absolute bottom-0 left-0 right-0 p-5">
                <p className="text-xs uppercase tracking-[0.2em] text-stone-300">Featured</p>
                <h2 className="mt-2 font-display text-3xl">{featured.name}</h2>
                <p className="mt-2 max-w-xl text-sm text-stone-200">{featured.description}</p>
                <div className="mt-4 flex items-center gap-2 text-xs text-stone-300">
                  <span>★ {featured.rating.toFixed(1)}</span>
                  <span>•</span>
                  <span>{featured.players}</span>
                  <span>•</span>
                  <span>{featured.price === 0 ? 'Free' : `$${featured.price}`}</span>
                </div>
              </div>
            </div>

            <div className="rounded-2xl border border-stone-200 bg-white p-4">
              <label htmlFor="store-search" className="mb-2 block text-xs uppercase tracking-[0.18em] text-stone-500">
                Advanced lookup
              </label>
              <div className="relative">
                <input
                  id="store-search"
                  value={query}
                  onChange={(event) => setQuery(event.target.value)}
                  placeholder="Search game, tag, studio..."
                  className="w-full rounded-xl border border-stone-300 bg-stone-50 px-4 py-3 pr-20 text-sm outline-none transition focus:border-stone-500 focus:ring-2 focus:ring-stone-200"
                />
                <span className="absolute right-2 top-2 rounded-lg border border-stone-300 bg-white px-2 py-1 text-[11px] text-stone-500">Ctrl/⌘ K</span>
              </div>

              <div className="mt-4 grid grid-cols-2 gap-2 text-sm">
                <label className="flex items-center gap-2 rounded-xl border border-stone-200 bg-stone-50 px-3 py-2">
                  <input type="checkbox" checked={freeOnly} onChange={() => setFreeOnly((value) => !value)} />
                  Free only
                </label>
                <label className="flex items-center gap-2 rounded-xl border border-stone-200 bg-stone-50 px-3 py-2">
                  <input type="checkbox" checked={installedOnly} onChange={() => setInstalledOnly((value) => !value)} />
                  Installed
                </label>
              </div>

              <div className="mt-4 grid gap-2">
                <select
                  value={mode}
                  onChange={(event) => setMode(event.target.value as (typeof modes)[number])}
                  className="rounded-xl border border-stone-300 bg-white px-3 py-2 text-sm"
                >
                  {modes.map((option) => (
                    <option key={option}>{option}</option>
                  ))}
                </select>
                <select
                  value={sortBy}
                  onChange={(event) => setSortBy(event.target.value as (typeof sortOptions)[number])}
                  className="rounded-xl border border-stone-300 bg-white px-3 py-2 text-sm"
                >
                  {sortOptions.map((option) => (
                    <option key={option}>{option}</option>
                  ))}
                </select>
              </div>
            </div>
          </div>
        </header>

        <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
          <div className="flex flex-wrap gap-2">
            {categories.map((item) => {
              const active = category === item
              return (
                <button
                  key={item}
                  onClick={() => setCategory(item)}
                  className={`rounded-full border px-4 py-2 text-sm transition ${
                    active
                      ? 'border-stone-800 bg-stone-900 text-white'
                      : 'border-stone-300 bg-white text-stone-600 hover:border-stone-500'
                  }`}
                >
                  {item}
                </button>
              )
            })}
          </div>
          <p className="text-sm text-stone-600">{filteredGames.length} results</p>
        </div>

        <section className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
          {filteredGames.map((game) => (
            <article
              key={game.id}
              className="group overflow-hidden rounded-2xl border border-stone-200 bg-white shadow-sm transition hover:-translate-y-0.5 hover:shadow-md"
            >
              <button type="button" onClick={() => setSelectedGame(game)} className="block w-full text-left">
                <div className="relative">
                  <img src={game.cover} alt={game.name} className="h-44 w-full object-cover" />
                  <div className="absolute inset-0 bg-gradient-to-t from-black/60 via-transparent to-transparent" />
                  <span className="absolute bottom-2 right-2 rounded-full bg-white/90 px-2 py-1 text-xs font-medium text-stone-700">
                    {game.price === 0 ? 'Free' : `$${game.price}`}
                  </span>
                </div>
                <div className="space-y-3 p-4">
                  <div>
                    <div className="flex items-center justify-between gap-2">
                      <h3 className="font-display text-xl">{game.name}</h3>
                      <span className="text-sm text-amber-600">★ {game.rating.toFixed(1)}</span>
                    </div>
                    <p className="mt-1 text-xs uppercase tracking-[0.16em] text-stone-500">{game.studio}</p>
                  </div>
                  <p className="min-h-10 text-sm text-stone-600">{game.description}</p>
                  <div className="flex flex-wrap gap-2">
                    {game.tags.slice(0, 3).map((tag) => (
                      <span key={tag} className="rounded-full bg-stone-100 px-2 py-1 text-xs text-stone-600">
                        #{tag}
                      </span>
                    ))}
                  </div>
                </div>
              </button>
            </article>
          ))}
        </section>
      </div>

      {selectedGame ? (
        <aside className="fixed inset-0 z-30 flex items-end bg-black/50 p-0 sm:items-center sm:justify-center sm:p-6" onClick={() => setSelectedGame(null)}>
          <div
            className="max-h-[90vh] w-full overflow-auto rounded-t-3xl border border-stone-200 bg-white sm:max-w-3xl sm:rounded-3xl"
            onClick={(event) => event.stopPropagation()}
          >
            <div className="relative">
              <img src={selectedGame.cover} alt={selectedGame.name} className="h-56 w-full object-cover" />
              <button
                onClick={() => setSelectedGame(null)}
                className="absolute right-3 top-3 rounded-full bg-white/90 px-3 py-1 text-sm text-stone-700"
              >
                Close
              </button>
            </div>
            <div className="p-5 sm:p-6">
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <h3 className="font-display text-3xl">{selectedGame.name}</h3>
                  <p className="mt-1 text-sm text-stone-500">{selectedGame.studio}</p>
                </div>
                <div className="text-right text-sm text-stone-600">
                  <p>★ {selectedGame.rating.toFixed(1)}</p>
                  <p>{selectedGame.players}</p>
                </div>
              </div>

              <p className="mt-4 text-stone-700">{selectedGame.description}</p>

              <div className="mt-4 flex flex-wrap gap-2">
                {selectedGame.modes.map((item) => (
                  <span key={item} className="rounded-full border border-stone-300 px-3 py-1 text-xs text-stone-600">
                    {item}
                  </span>
                ))}
              </div>

              <div className="mt-6 flex flex-wrap gap-3">
                <button className="rounded-xl bg-stone-900 px-5 py-3 text-sm font-medium text-white hover:bg-stone-800">
                  {selectedGame.installed ? 'Play Now' : selectedGame.price === 0 ? 'Install Free' : `Buy $${selectedGame.price}`}
                </button>
                <button className="rounded-xl border border-stone-300 px-5 py-3 text-sm font-medium text-stone-700 hover:border-stone-500">
                  Add to Collection
                </button>
              </div>
            </div>
          </div>
        </aside>
      ) : null}
    </div>
  )
}

export default App
