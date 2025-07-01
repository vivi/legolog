import math
import yaml

LOG_LIFETIME = 31
HASH_SIZE_BYTES = 16


class DeveloperParams:
    def __init__(self, file_path): 
        with open(file_path, "r") as f:
            params = yaml.load(f, Loader=yaml.Loader)
        self.verify_period = params['writer_refresh_rate']
        self.writer_refresh_rate = params['writer_refresh_rate']

        self.log_size = params['log_size']
        self.max_offline_time = params['writer_max_offline_time']

        self.max_read_latency = params['max_read_latency']
        self.max_read_throughput = params['max_read_throughput']  / self.max_read_latency

        self.max_write_latency = params['max_write_latency']
        self.max_write_throughput = params['max_write_throughput'] / self.max_write_latency

        self.writer_device = params['writer_device']
        self.reader_device = params['reader_device']
        self.auditor_device = params['auditor_device']
    
    def determine_agg_history(self):
        if self.max_offline_time <= self.verify_period:
            return False
        if self.writer_device == "mobile":
            return True
        if self.max_read_throughput / self.max_write_throughput > 10:
            return False
        else:
            return True
         
    def determine_monitor(self):
        if self.auditor_device == "server":
            return True

        
class SystemParams:
    def __init__(self, log_size: int, verify_period: float, update_period: float, num_partitions: int, agg_history: bool, dev_params: DeveloperParams):
        self.log_size = log_size
        self.verify_period = verify_period
        self.update_period = update_period
        self.num_partitions = num_partitions
        self.agg_history = agg_history

        self.read_tput = dev_params.max_read_throughput
        self.update_tput = dev_params.max_write_throughput
        self.updates_per_update_period = self.update_tput * self.update_period
        self.max_offline_time = dev_params.max_offline_time

        self.dev_params = dev_params
    
    def calc_update_periods_per_verification_period(self):
        return self.verify_period / self.update_period

    def calc_monitor_proof_hashes(self):
        """Returns total hashes computed by clients in the system in a single max_offline_time."""
        if self.agg_history:
            return self.calc_base_tree_proof_hashes()
        else:  
            verify_periods_to_monitor = math.floor(self.max_offline_time / self.verify_period)
            return verify_periods_to_monitor * self.calc_base_tree_proof_hashes()
    
    def calc_monitor_proof_bytes(self):
        return self.calc_monitor_proof_hashes() * HASH_SIZE_BYTES

    def calc_base_tree_proof_hashes(self):
        return 1 + max(0, math.ceil(math.log2(self.log_size / self.num_partitions)))

    def calc_lookup_proof_hashes(self):
        """Returns the number of hashes the client must compute for a single lookup."""
        update_tree_proof_size = (1 + math.ceil(max(1, math.log2(self.updates_per_update_period / self.num_partitions))))  * self.calc_update_periods_per_verification_period()
        base_tree_proof_size = 1 + math.ceil(math.log2(self.log_size / self.num_partitions))
        if self.agg_history:
            base_tree_proof_size = base_tree_proof_size * LOG_LIFETIME
        return update_tree_proof_size + base_tree_proof_size

    def calc_lookup_proof_bytes(self):
        return self.calc_lookup_proof_hashes() * HASH_SIZE_BYTES

    def calc_auditor_proof_hashes(self):
        """Returns the number of hashes the auditor must process every update period."""
        base_tree_proof_size = 1
        if self.agg_history:
            base_tree_proof_size = LOG_LIFETIME
        base_tree_proof_size /= self.calc_update_periods_per_verification_period()
        update_tree_proof_size = math.ceil(math.log2(self.calc_update_periods_per_verification_period())) 
        return self.num_partitions * (base_tree_proof_size + update_tree_proof_size)

    def calc_digest_bw(self):
        base_tree_proof_size = 1
        if self.agg_history:
            base_tree_proof_size = LOG_LIFETIME
        update_log_proof_size = 1
        return (base_tree_proof_size + update_log_proof_size) * HASH_SIZE_BYTES * self.read_tput

    def calc_auditor_proof_bytes(self):
        return self.calc_auditor_proof_hashes() * HASH_SIZE_BYTES
    
    def calc_work(self):
        res = 0
        if self.dev_params.reader_device == "mobile":
            res += self.calc_read_work()
        if self.dev_params.writer_device == "mobile":
            res += self.calc_monitor_work()
        if self.dev_params.auditor_device == "mobile":
            res += self.calc_auditor_proof_hashes()
        if self.dev_params.reader_device != "mobile" and self.dev_params.writer_device != "mobile" and self.dev_params.auditor_device != "mobile":
            res = self.calc_read_work() + self.calc_monitor_work() + self.calc_auditor_proof_hashes() / self.update_period
        return res
        
    def calc_read_work(self):
        return self.calc_lookup_proof_hashes() * self.read_tput

    def calc_monitor_work(self):
        return self.calc_monitor_proof_hashes() * self.log_size / self.max_offline_time

    def calc_client_work(self):
        #return self.calc_monitor_proof_hashes() * self.log_size / self.max_offline_time
        return self.calc_lookup_proof_hashes() * self.read_tput + self.calc_monitor_proof_hashes() * self.log_size / self.max_offline_time + self.calc_auditor_proof_hashes() / self.update_period
        return self.calc_lookup_proof_hashes() * self.read_tput + (self.calc_monitor_proof_hashes() * self.log_size) / self.max_offline_time
    
    def calc_server_requirements(self):
        per_update_period = self.calc_lookup_proof_hashes() * self.read_tput * self.update_period   \
            + self.calc_monitor_proof_hashes() * self.log_size / self.max_offline_time * self.update_period \
            + self.calc_auditor_proof_hashes()
        return per_update_period

        #
        #self.calc_lookup_proof_hashes() * self.read_tput + self.calc_monitor_proof_hashes() * self.log_size / self.max_offline_time + self.calc_auditor_proof_hashes() / self.update_period
        # how many hashes do you need to crunch updates in a verification period?

def calc_verify(params: DeveloperParams, min_verify_period, max_verify_period, num_partitions=1, agg_history=False):
    curr_params = SystemParams(
        log_size=params.log_size,
        verify_period=min_verify_period,
        update_period=params.max_write_latency,
        num_partitions=num_partitions,
        agg_history=agg_history,
        dev_params=params
    )
    best_verify_period = 0
    min_client_work = 0
    for i in range(min_verify_period, max_verify_period+1, params.max_write_latency):#params.max_write_latency):
        curr_params.verify_period = i
        client_work = curr_params.calc_work()
        if best_verify_period == 0 or client_work < min_client_work:
            best_verify_period = i
            min_client_work = client_work
            #print(f"{agg_history} {num_partitions}: {min_verify_period}, {min_client_work}")
    #print(f"{agg_history} {num_partitions}: {min_verify_period}, {min_client_work}")
    return best_verify_period, min_client_work


def calc_optimal_params(dev_params: DeveloperParams, agghist: bool = False, always_prevent_misbehavior: bool = False): 
    min_partition = 0
    min_verify_period = 0
    min_client_work = 0
    for i in range(10, 1000000, 100):
        if always_prevent_misbehavior:
            v, m = calc_verify(dev_params, dev_params.max_offline_time, dev_params.max_offline_time, i, agghist)
        else:
            v, m = calc_verify(dev_params, dev_params.max_write_latency, dev_params.max_offline_time, i, agghist)
        if min_verify_period == 0 or m < min_client_work:
            params = SystemParams(
                log_size=dev_params.log_size,
                verify_period=v,
                update_period=dev_params.max_write_latency,
                num_partitions=i,
                agg_history=agghist,
                dev_params=dev_params
            )
            if params.calc_auditor_proof_hashes() > 1e5:
                break
            min_partition = i
            min_verify_period = v
            min_client_work = m

    params = SystemParams(
        log_size=dev_params.log_size,
        verify_period=min_verify_period,
        update_period=dev_params.max_write_latency,
        num_partitions=min_partition,
        agg_history=agghist,
        dev_params=dev_params
    )
    print(f"{min_partition}, {min_verify_period}, {min_client_work:.1f}, {params.calc_auditor_proof_hashes():.1f}, {params.calc_server_requirements() / params.update_period :.1f}")

#dev_params = DeveloperParams("developer_params.yaml")
dev_params = DeveloperParams("merkle2_kt.yaml")
dev_params = DeveloperParams("signal_kt_scaled.yaml")
dev_params = DeveloperParams("coniks_kt.yaml")
dev_params = DeveloperParams("binary_transparency.yaml")
#dev_params = DeveloperParams("force_agghist_params.yaml")

calc_optimal_params(dev_params, agghist=False, always_prevent_misbehavior=False)
calc_optimal_params(dev_params, agghist=True, always_prevent_misbehavior=False)
#calc_optimal_params(dev_params, agghist=False, always_prevent_misbehavior=True)
#calc_optimal_params(dev_params, agghist=True, always_prevent_misbehavior=True)
