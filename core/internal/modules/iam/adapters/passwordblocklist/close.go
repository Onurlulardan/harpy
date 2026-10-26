package passwordblocklist

import blocklistv1 "github.com/Onurlulardan/harpy/contracts/passwordblocklist/v1"

func (blocklist *Blocklist) Close() error {
	if blocklist == nil {
		return nil
	}
	blocklist.mu.Lock()
	defer blocklist.mu.Unlock()
	if !blocklist.open {
		return nil
	}
	clear(blocklist.content)
	blocklist.content = nil
	blocklist.artifact = blocklistv1.Artifact{}
	blocklist.open = false
	return nil
}
