package ssh

type AllListener struct {
	CPUInfoListener           Listener[CPUInfoMessage]
	CPUPerformanceListener    Listener[CPUPerformanceMessage]
	MemoryPerformanceListener Listener[MemoryPerformanceMessage]
	UptimeListener            Listener[UptimeMessage]
	LoadavgListener           Listener[LoadavgMessage]
	NetDevListener            Listener[NetDevMessage]
	NetStatListener           Listener[NetStatMessage]
	TempListener              Listener[TempMessage]
	DiskListener              Listener[DiskMessage]
	ProcessListener           Listener[ProcessMessage]
}
