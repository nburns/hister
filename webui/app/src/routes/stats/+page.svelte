<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { apiFetch } from '$lib/api';
  import { PageHeader } from '@hister/components';
  import { Badge } from '@hister/components/ui/badge';
  import * as Card from '@hister/components/ui/card';
  import * as Table from '@hister/components/ui/table';
  import * as Alert from '@hister/components/ui/alert';
  import AlertCircle from '@lucide/svelte/icons/circle-alert';
  import { Activity } from '@lucide/svelte';

  interface CrawlerURLCounts {
    pending: number;
    in_progress: number;
    done: number;
    failed: number;
    skipped: number;
    count_2xx: number;
    count_3xx: number;
    count_4xx: number;
    count_5xx: number;
  }

  interface CrawlerJob {
    id: string;
    start_url: string;
    label: string;
    status: 'running' | 'completed' | 'interrupted';
    pages_fetched: number;
    bytes_fetched: number;
    retries: number;
    breaker_trips: number;
    robots_denials: number;
    budget_stops: number;
    url_counts: CrawlerURLCounts;
    created_at: string;
    updated_at: string;
  }

  interface CrawlerDebugResponse {
    jobs: CrawlerJob[];
  }

  let jobs: CrawlerJob[] = $state([]);
  let error = $state('');
  let loading = $state(true);
  let pollTimer: ReturnType<typeof setTimeout> | undefined;

  function formatBytes(n: number): string {
    if (n >= 1_073_741_824) return (n / 1_073_741_824).toFixed(1) + ' GB';
    if (n >= 1_048_576) return (n / 1_048_576).toFixed(1) + ' MB';
    if (n >= 1024) return (n / 1024).toFixed(1) + ' KB';
    return n + ' B';
  }

  function formatRelative(iso: string): string {
    const diff = Math.floor((Date.now() - new Date(iso).getTime()) / 1000);
    if (diff < 60) return diff + 's ago';
    if (diff < 3600) return Math.floor(diff / 60) + 'm ago';
    if (diff < 86400) return Math.floor(diff / 3600) + 'h ago';
    return Math.floor(diff / 86400) + 'd ago';
  }

  function pollInterval(currentJobs: CrawlerJob[]): number {
    return currentJobs.some((j) => j.status === 'running') ? 5000 : 15000;
  }

  async function fetchJobs() {
    try {
      const res = await apiFetch('/stats/crawler');
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const data: CrawlerDebugResponse = await res.json();
      jobs = data.jobs ?? [];
      error = '';
    } catch (e) {
      error = String(e);
    } finally {
      loading = false;
    }
    schedulePoll();
  }

  function schedulePoll() {
    clearTimeout(pollTimer);
    pollTimer = setTimeout(() => fetchJobs(), pollInterval(jobs));
  }

  onMount(() => {
    void fetchJobs();
  });

  onDestroy(() => {
    clearTimeout(pollTimer);
  });
</script>

<svelte:head>
  <title>Hister - Crawl Stats</title>
</svelte:head>

<div class="flex flex-1 flex-col gap-8 overflow-y-auto px-6 py-8 md:gap-10 md:px-12 md:py-12">
  <div class="flex flex-col gap-4">
    <PageHeader color="hister-lime" size="lg" class="uppercase">Crawl Stats</PageHeader>
  </div>

  {#if error}
    <Alert.Root variant="error" class="shadow-brutal border-[3px]">
      <AlertCircle class="size-5 shrink-0" />
      <Alert.Description class="font-inter text-sm">{error}</Alert.Description>
    </Alert.Root>
  {/if}

  {#if loading}
    <div class="flex items-center justify-center py-16">
      <p class="font-inter text-text-brand-muted text-lg">Loading crawl stats...</p>
    </div>
  {:else}
    <Card.Root>
      <Card.Header color="hister-lime">
        <div class="flex h-12 w-12 shrink-0 items-center justify-center bg-white/20">
          <Activity class="size-6 text-white" />
        </div>
        <div class="flex flex-col gap-1">
          <Card.Title class="font-space text-xl font-extrabold tracking-[1px] text-white uppercase"
            >Crawl Jobs</Card.Title
          >
          <Card.Description class="font-inter text-sm text-white/80"
            >{jobs.length} job{jobs.length === 1 ? '' : 's'}</Card.Description
          >
        </div>
      </Card.Header>

      <Card.Content class="flex-1 p-0">
        {#if jobs.length === 0}
          <div class="flex flex-col items-center justify-center gap-3 py-10">
            <div
              class="flex h-12 w-12 items-center justify-center"
              style="background-color: color-mix(in srgb, var(--hister-lime) 10%, transparent); color: var(--hister-lime);"
            >
              <Activity class="size-5" />
            </div>
            <p class="font-inter text-text-brand-muted text-sm">No crawl jobs yet.</p>
          </div>
        {:else}
          <div class="overflow-x-auto">
            <Table.Root>
              <Table.Header>
                <Table.Row
                  class="bg-muted-surface border-brutal-border hover:bg-muted-surface border-b-[3px]"
                >
                  <Table.Head
                    class="font-space text-text-brand-muted h-auto px-2 py-3 text-xs font-bold tracking-[1px] uppercase md:px-5"
                    >Job ID</Table.Head
                  >
                  <Table.Head
                    class="font-space text-text-brand-muted h-auto px-2 py-3 text-xs font-bold tracking-[1px] uppercase md:px-5"
                    >Start URL</Table.Head
                  >
                  <Table.Head
                    class="font-space text-text-brand-muted h-auto px-2 py-3 text-xs font-bold tracking-[1px] uppercase md:px-5"
                    >Status</Table.Head
                  >
                  <Table.Head
                    class="font-space text-text-brand-muted h-auto px-2 py-3 text-xs font-bold tracking-[1px] uppercase md:px-5"
                    >Pages</Table.Head
                  >
                  <Table.Head
                    class="font-space text-text-brand-muted h-auto px-2 py-3 text-xs font-bold tracking-[1px] uppercase md:px-5"
                    >Bytes</Table.Head
                  >
                  <Table.Head
                    class="font-space text-text-brand-muted h-auto px-2 py-3 text-xs font-bold tracking-[1px] uppercase md:px-5"
                    >HTTP</Table.Head
                  >
                  <Table.Head
                    class="font-space text-text-brand-muted h-auto px-2 py-3 text-xs font-bold tracking-[1px] uppercase md:px-5"
                    >Retries</Table.Head
                  >
                  <Table.Head
                    class="font-space text-text-brand-muted h-auto px-2 py-3 text-xs font-bold tracking-[1px] uppercase md:px-5"
                    >Robots</Table.Head
                  >
                  <Table.Head
                    class="font-space text-text-brand-muted h-auto px-2 py-3 text-xs font-bold tracking-[1px] uppercase md:px-5"
                    >Updated</Table.Head
                  >
                </Table.Row>
              </Table.Header>
              <Table.Body>
                {#each jobs as job (job.id)}
                  <Table.Row class="border-brutal-border border-b-[3px]">
                    <Table.Cell class="font-fira text-text-brand px-2 py-3 text-xs md:px-5">
                      <span title={job.id}>{job.id.slice(0, 8)}</span>
                    </Table.Cell>
                    <Table.Cell
                      class="font-fira text-text-brand-secondary max-w-48 truncate px-2 py-3 text-sm md:px-5"
                    >
                      <span title={job.start_url}>{job.start_url}</span>
                    </Table.Cell>
                    <Table.Cell class="px-2 py-3 md:px-5">
                      <Badge
                        variant="default"
                        class="font-space border-0 px-2 py-1 text-xs font-bold tracking-[0.5px] uppercase {job.status ===
                        'running'
                          ? 'bg-hister-lime text-white'
                          : job.status === 'completed'
                            ? 'bg-muted-surface text-text-brand-secondary'
                            : 'bg-hister-amber text-white'}"
                      >
                        {job.status}
                      </Badge>
                    </Table.Cell>
                    <Table.Cell class="font-inter text-text-brand px-2 py-3 text-sm md:px-5">
                      {job.pages_fetched}
                    </Table.Cell>
                    <Table.Cell class="font-inter text-text-brand px-2 py-3 text-sm md:px-5">
                      {formatBytes(job.bytes_fetched)}
                    </Table.Cell>
                    <Table.Cell class="px-2 py-3 md:px-5">
                      <div class="flex flex-wrap gap-1">
                        {#if job.url_counts.count_2xx > 0}
                          <span
                            class="font-fira rounded-none bg-hister-lime/20 px-1.5 py-0.5 text-xs text-hister-lime"
                            >2xx:{job.url_counts.count_2xx}</span
                          >
                        {/if}
                        {#if job.url_counts.count_3xx > 0}
                          <span
                            class="font-fira rounded-none bg-hister-indigo/20 px-1.5 py-0.5 text-xs text-hister-indigo"
                            >3xx:{job.url_counts.count_3xx}</span
                          >
                        {/if}
                        {#if job.url_counts.count_4xx > 0}
                          <span
                            class="font-fira rounded-none bg-hister-amber/20 px-1.5 py-0.5 text-xs text-hister-amber"
                            >4xx:{job.url_counts.count_4xx}</span
                          >
                        {/if}
                        {#if job.url_counts.count_5xx > 0}
                          <span
                            class="font-fira rounded-none bg-hister-rose/20 px-1.5 py-0.5 text-xs text-hister-rose"
                            >5xx:{job.url_counts.count_5xx}</span
                          >
                        {/if}
                      </div>
                    </Table.Cell>
                    <Table.Cell class="font-inter text-text-brand px-2 py-3 text-sm md:px-5">
                      {job.retries}
                    </Table.Cell>
                    <Table.Cell class="font-inter text-text-brand px-2 py-3 text-sm md:px-5">
                      {job.robots_denials}
                    </Table.Cell>
                    <Table.Cell
                      class="font-inter text-text-brand-muted px-2 py-3 text-sm md:px-5"
                      title={job.updated_at}
                    >
                      {formatRelative(job.updated_at)}
                    </Table.Cell>
                  </Table.Row>
                {/each}
              </Table.Body>
            </Table.Root>
          </div>
        {/if}
      </Card.Content>
    </Card.Root>
  {/if}
</div>
