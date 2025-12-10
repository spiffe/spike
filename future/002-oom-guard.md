Psedocode; to be added to the SDK and used within SPIKE components.

add oom guard

      package memory

      const CgroupPath = "/sys/fs/cgroup/"
      const MemLimitFile = "memory/memory.limit_in_bytes"
      const MemUsageFile = "memory/memory.usage_in_bytes"
      const MemMaxFile = "memory.max"
      const MemCurrentFile = "memory.current"


      type Watcher struct {
      memMax uint64
      memCurrentPath string
      memThreshold uint8
      interval time.Duration
      ctx context.Context
      cancel context.CancelFunc
      }

      func readUint(path string) (uint64, error) {
      b, err := os.ReadFile(path)
      if err != nil {
      return 0, err
      }
      return strconv.ParseUint(strings.TrimSpace(string(b)), 10, 64)
      }

      func discover(memMaxPath, memCurrentPath string) (string, string, error) {
      if memMaxPath == "" {
      maxPathV1 := filepath.Join(CgroupPath, MemLimitFile)
      maxPathV2 := filepath.Join(CgroupPath, MemMaxFile)

      if _, err := os.Lstat(maxPathV2); err == nil {
      memMaxPath = maxPathV2
      } else if _, err = os.Lstat(maxPathV1); err == nil {
      memMaxPath = maxPathV1
      }
      }
      if memCurrentPath == "" {
      currentPathV1 := filepath.Join(CgroupPath, MemUsageFile)
      currentPathV2 := filepath.Join(CgroupPath, MemCurrentFile)
      if _, err := os.Lstat(currentPathV2); err == nil {
      memCurrentPath = currentPathV2
      } else if _, err = os.Lstat(currentPathV1); err == nil {
      memCurrentPath = currentPathV1
      }
      }

      if memMaxPath == "" && memCurrentPath == "" {
      err
      }
      if memMaxPath == "" {
      err
      }
      if memCurrentPath == "" {
      err
      }

      return memMaxPath, memCurrentPath, nil
      }

      func (w *Watcher) run(ctx context.Context) {
      t := time.NewTicker(w.interval)
      defer t.Stop()

      for {
      select {
      case <-ctx.Done():
      log
      return
      case <-t.C:
      current, err := readUint(w.memCurrentPath)
      if err != nil {
      log
      continue
      }

      currentPercentage := float64(current) / float64(w.memMax) * 100
      if currentPercentage >= float64(w.memThreshold) {
      log
      w.cancel()
      return
      }
      log
      }
      }
      }

      func New(memMaxPath, memCurrentPath string, memThreshold uint8, interval time.Duration) (*Watcher, error) {
      if memThreshold < 1 || memThreshold > 100 {
      return nil, err
      }

      if minInterval := 50 * time.Millisecond; interval < minInterval {
      return nil, err
      }

      memMaxPath, memCurrentPath, err = discover(memMaxPath, memCurrentPath)
      if err != nil {
      return nil, err
      }

      if _, err = os.Lstat(memCurrentPath); err != nil {
      return nil, err
      }

      memMax, err := readUint(memMaxPath)
      if err != nil {
      return nil, err
      }

      return &Watcher{
      memMax: memMax,
      memCurrentPath: memCurrentPath,
      memThreshold: memThreshold,
      interval: interval,
      }, nil
      }

      func (w *Watcher) Watch(ctx context.Context) context.Context {
      sync.Once.Do(func() {
      w.ctx, w.cancel = context.WithCancel(ctx)
      go w.run(ctx)
      })
      return w.ctx
      }