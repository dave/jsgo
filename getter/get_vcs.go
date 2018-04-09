package getter

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"regexp"

	"net/url"

	"context"

	"github.com/dave/jsgo/gitcache"
	"gopkg.in/src-d/go-billy.v4"
)

var defaultSecureScheme = map[string]bool{
	"https":   true,
	"git+ssh": true,
	"bzr+ssh": true,
	"svn+ssh": true,
	"ssh":     true,
}

func isSecure(repo string) bool {
	u, err := url.Parse(repo)
	if err != nil {
		// If repo is not a URL, it's not secure.
		return false
	}
	return defaultSecureScheme[u.Scheme]
}

// vcsList lists the known version control systems
var vcsList = []string{
	"git",
}

func vcsByCmd(cmd string, gitreq *gitcache.Request) vcsProvider {
	switch cmd {
	case "git":
		return &gitProvider{gitreq: gitreq}
	}
	return nil
}

func (r *repoRoot) create(ctx context.Context, fs billy.Filesystem) error {
	if err := r.vcs.create(ctx, r.repo, r.dir, fs); err != nil {
		return err
	}
	r.exists = true
	r.hash = r.vcs.hash()
	return nil
}

func (r *repoRoot) download(ctx context.Context) error {
	if err := r.vcs.download(ctx); err != nil {
		return err
	}
	r.hash = r.vcs.hash()
	return nil
}

// A vcsPath describes how to convert an import path into a
// version control system and repository name.
type vcsPath struct {
	prefix string                                                                             // prefix this description applies to
	re     string                                                                             // pattern for import path
	repo   string                                                                             // repository to use (expand with match of re)
	vcs    string                                                                             // version control system to use (expand with match of re)
	check  func(ctx context.Context, match map[string]string, gitreq *gitcache.Request) error // additional checks
	ping   bool                                                                               // ping for scheme to use to download repo

	regexp *regexp.Regexp // cached compiled form of re
}

// vcsFromDir inspects dir and its parents to determine the
// version control system and code repository to use.
// On return, root is the import path
// corresponding to the root of the repository.
func (g *Getter) vcsFromDir(dir, srcRoot string) (root *repoRoot, err error) {
	// Clean and double-check that dir is in (a subdirectory of) srcRoot.
	dir = filepath.Clean(dir)
	srcRoot = filepath.Clean(srcRoot)
	if len(dir) <= len(srcRoot) || dir[len(srcRoot)] != filepath.Separator {
		return nil, fmt.Errorf("directory %q is outside source root %q", dir, srcRoot)
	}

	origDir := dir
	for len(dir) > len(srcRoot) {
		if root, ok := g.downloadRootCache[dir]; ok {
			return root, nil
		}

		// Move to parent.
		ndir := filepath.Dir(dir)
		if len(ndir) >= len(dir) {
			// Shouldn't happen, but just in case, stop.
			break
		}
		dir = ndir
	}

	return nil, fmt.Errorf("directory %q is not using a known version control system", origDir)
}

type repoRoot struct {
	vcs vcsProvider

	// repo is the repository URL, including scheme
	repo string

	// root is the import path corresponding to the root of the
	// repository
	root string

	// isCustom is true for custom import paths (those defined by HTML meta tags)
	isCustom bool

	// dir is the dir corresponding to the root of the repository
	dir string

	// exists is true once the repo has been downloaded
	exists bool

	// hash is the latest commit hash
	hash string
}

var httpPrefixRE = regexp.MustCompile(`^https?:`)

// repoRootForImportPath analyzes importPath to determine the
// version control system, and code repository to use.
func (g *Getter) repoRootForImportPath(ctx context.Context, importPath string, insecure bool) (*repoRoot, error) {
	rr, err := g.repoRootFromVCSPaths(ctx, importPath, "", insecure, vcsPaths)
	if err == errUnknownSite {
		rr, err = g.repoRootForImportDynamic(ctx, importPath, insecure)
		if err != nil {
			err = fmt.Errorf("unrecognized import path %q (%v)", importPath, err)
		}
	}
	if err != nil {
		rr1, err1 := g.repoRootFromVCSPaths(ctx, importPath, "", insecure, vcsPathsAfterDynamic)
		if err1 == nil {
			rr = rr1
			err = nil
		}
	}

	if err == nil && strings.Contains(importPath, "...") && strings.Contains(rr.root, "...") {
		// Do not allow wildcards in the repo root.
		rr = nil
		err = fmt.Errorf("cannot expand ... in %q", importPath)
	}
	return rr, err
}

var errUnknownSite = errors.New("dynamic lookup required to find mapping")

// repoRootFromVCSPaths attempts to map importPath to a repoRoot
// using the mappings defined in vcsPaths.
// If scheme is non-empty, that scheme is forced.
func (g *Getter) repoRootFromVCSPaths(ctx context.Context, importPath, scheme string, insecure bool, vcsPaths []*vcsPath) (*repoRoot, error) {
	// A common error is to use https://packagepath because that's what
	// hg and git require. Diagnose this helpfully.
	if loc := httpPrefixRE.FindStringIndex(importPath); loc != nil {
		// The importPath has been cleaned, so has only one slash. The pattern
		// ignores the slashes; the error message puts them back on the RHS at least.
		return nil, fmt.Errorf("%q not allowed in import path", importPath[loc[0]:loc[1]]+"//")
	}
	for _, srv := range vcsPaths {
		if !strings.HasPrefix(importPath, srv.prefix) {
			continue
		}
		m := srv.regexp.FindStringSubmatch(importPath)
		if m == nil {
			if srv.prefix != "" {
				return nil, fmt.Errorf("invalid %s import path %q", srv.prefix, importPath)
			}
			continue
		}

		// Build map of named subexpression matches for expand.
		match := map[string]string{
			"prefix": srv.prefix,
			"import": importPath,
		}
		for i, name := range srv.regexp.SubexpNames() {
			if name != "" && match[name] == "" {
				match[name] = m[i]
			}
		}
		if srv.vcs != "" {
			match["vcs"] = expand(match, srv.vcs)
		}
		if srv.repo != "" {
			match["repo"] = expand(match, srv.repo)
		}
		if srv.check != nil {
			if err := srv.check(ctx, match, g.gitreq); err != nil {
				return nil, err
			}
		}
		vcs := vcsByCmd(match["vcs"], g.gitreq)
		if vcs == nil {
			return nil, fmt.Errorf("unknown version control system %q", match["vcs"])
		}
		if srv.ping {
			if scheme != "" {
				match["repo"] = scheme + "://" + match["repo"]
			} else {
				for _, scheme := range vcs.schemes() {
					if !insecure && !defaultSecureScheme[scheme] {
						continue
					}
					if vcs.ping(ctx, scheme, match["repo"]) == nil {
						match["repo"] = scheme + "://" + match["repo"]
						break
					}
				}
			}
		}
		rr := &repoRoot{
			vcs:  vcs,
			repo: match["repo"],
			root: match["root"],
		}
		return rr, nil
	}
	return nil, errUnknownSite
}

// repoRootForImportDynamic finds a *repoRoot for a custom domain that's not
// statically known by repoRootForImportPathStatic.
//
// This handles custom import paths like "name.tld/pkg/foo" or just "name.tld".
func (g *Getter) repoRootForImportDynamic(ctx context.Context, importPath string, insecure bool) (*repoRoot, error) {
	slash := strings.Index(importPath, "/")
	if slash < 0 {
		slash = len(importPath)
	}
	host := importPath[:slash]
	if !strings.Contains(host, ".") {
		return nil, errors.New("import path does not begin with hostname")
	}
	urlStr, body, err := GetMaybeInsecure(ctx, importPath, insecure)
	if err != nil {
		msg := "https fetch: %v"
		if insecure {
			msg = "http/" + msg
		}
		return nil, fmt.Errorf(msg, err)
	}
	defer body.Close()
	imports, err := parseMetaGoImports(body)
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %v", importPath, err)
	}
	// Find the matched meta import.
	mmi, err := matchGoImport(imports, importPath)
	if err != nil {
		if _, ok := err.(ImportMismatchError); !ok {
			return nil, fmt.Errorf("parse %s: %v", urlStr, err)
		}
		return nil, fmt.Errorf("parse %s: no go-import meta tags (%s)", urlStr, err)
	}
	// If the import was "uni.edu/bob/project", which said the
	// prefix was "uni.edu" and the RepoRoot was "evilroot.com",
	// make sure we don't trust Bob and check out evilroot.com to
	// "uni.edu" yet (possibly overwriting/preempting another
	// non-evil student). Instead, first verify the root and see
	// if it matches Bob's claim.
	if mmi.Prefix != importPath {
		urlStr0 := urlStr
		var imports []metaImport
		urlStr, imports, err = g.metaImportsForPrefix(ctx, mmi.Prefix, insecure)
		if err != nil {
			return nil, err
		}
		metaImport2, err := matchGoImport(imports, importPath)
		if err != nil || mmi != metaImport2 {
			return nil, fmt.Errorf("%s and %s disagree about go-import for %s", urlStr0, urlStr, mmi.Prefix)
		}
	}

	if err := validateRepoRoot(mmi.RepoRoot); err != nil {
		return nil, fmt.Errorf("%s: invalid repo root %q: %v", urlStr, mmi.RepoRoot, err)
	}
	rr := &repoRoot{
		vcs:      vcsByCmd(mmi.VCS, g.gitreq),
		repo:     mmi.RepoRoot,
		root:     mmi.Prefix,
		isCustom: true,
	}
	if rr.vcs == nil {
		return nil, fmt.Errorf("%s: unknown vcs %q", urlStr, mmi.VCS)
	}
	return rr, nil
}

// validateRepoRoot returns an error if repoRoot does not seem to be
// a valid URL with scheme.
func validateRepoRoot(repoRoot string) error {
	url, err := url.Parse(repoRoot)
	if err != nil {
		return err
	}
	if url.Scheme == "" {
		return errors.New("no scheme")
	}
	return nil
}

// metaImportsForPrefix takes a package's root import path as declared in a <meta> tag
// and returns its HTML discovery URL and the parsed metaImport lines
// found on the page.
//
// The importPath is of the form "golang.org/x/tools".
// It is an error if no imports are found.
// urlStr will still be valid if err != nil.
// The returned urlStr will be of the form "https://golang.org/x/tools?go-get=1"
func (g *Getter) metaImportsForPrefix(ctx context.Context, importPrefix string, insecure bool) (urlStr string, imports []metaImport, err error) {
	setCache := func(res fetchResult) (fetchResult, error) {
		g.fetchCacheMu.Lock()
		defer g.fetchCacheMu.Unlock()
		g.fetchCache[importPrefix] = res
		return res, nil
	}

	resi, _, _ := g.fetchGroup.Do(importPrefix, func() (resi interface{}, err error) {
		g.fetchCacheMu.Lock()
		if res, ok := g.fetchCache[importPrefix]; ok {
			g.fetchCacheMu.Unlock()
			return res, nil
		}
		g.fetchCacheMu.Unlock()

		urlStr, body, err := GetMaybeInsecure(ctx, importPrefix, insecure)
		if err != nil {
			return setCache(fetchResult{urlStr: urlStr, err: fmt.Errorf("fetch %s: %v", urlStr, err)})
		}
		imports, err := parseMetaGoImports(body)
		if err != nil {
			return setCache(fetchResult{urlStr: urlStr, err: fmt.Errorf("parsing %s: %v", urlStr, err)})
		}
		if len(imports) == 0 {
			err = fmt.Errorf("fetch %s: no go-import meta tag", urlStr)
		}
		return setCache(fetchResult{urlStr: urlStr, imports: imports, err: err})
	})
	res := resi.(fetchResult)
	return res.urlStr, res.imports, res.err
}

type fetchResult struct {
	urlStr  string // e.g. "https://foo.com/x/bar?go-get=1"
	imports []metaImport
	err     error
}

// metaImport represents the parsed <meta name="go-import"
// content="prefix vcs reporoot" /> tags from HTML files.
type metaImport struct {
	Prefix, VCS, RepoRoot string
}

func splitPathHasPrefix(path, prefix []string) bool {
	if len(path) < len(prefix) {
		return false
	}
	for i, p := range prefix {
		if path[i] != p {
			return false
		}
	}
	return true
}

// A ImportMismatchError is returned where metaImport/s are present
// but none match our import path.
type ImportMismatchError struct {
	importPath string
	mismatches []string // the meta imports that were discarded for not matching our importPath
}

func (m ImportMismatchError) Error() string {
	formattedStrings := make([]string, len(m.mismatches))
	for i, pre := range m.mismatches {
		formattedStrings[i] = fmt.Sprintf("meta tag %s did not match import path %s", pre, m.importPath)
	}
	return strings.Join(formattedStrings, ", ")
}

// matchGoImport returns the metaImport from imports matching importPath.
// An error is returned if there are multiple matches.
// errNoMatch is returned if none match.
func matchGoImport(imports []metaImport, importPath string) (metaImport, error) {
	match := -1
	imp := strings.Split(importPath, "/")

	errImportMismatch := ImportMismatchError{importPath: importPath}
	for i, im := range imports {
		pre := strings.Split(im.Prefix, "/")

		if !splitPathHasPrefix(imp, pre) {
			errImportMismatch.mismatches = append(errImportMismatch.mismatches, im.Prefix)
			continue
		}

		if match != -1 {
			return metaImport{}, fmt.Errorf("multiple meta tags match import path %q", importPath)
		}
		match = i
	}

	if match == -1 {
		return metaImport{}, errImportMismatch
	}
	return imports[match], nil
}

// expand rewrites s to replace {k} with match[k] for each key k in match.
func expand(match map[string]string, s string) string {
	for k, v := range match {
		s = strings.Replace(s, "{"+k+"}", v, -1)
	}
	return s
}

// vcsPaths defines the meaning of import paths referring to
// commonly-used VCS hosting sites (github.com/user/dir)
// and import paths referring to a fully-qualified importPath
// containing a VCS type (foo.com/repo.git/dir)
var vcsPaths = []*vcsPath{
	// Github
	{
		prefix: "github.com/",
		re:     `^(?P<root>github\.com/[A-Za-z0-9_.\-]+/[A-Za-z0-9_.\-]+)(/[\p{L}0-9_.\-]+)*$`,
		vcs:    "git",
		repo:   "https://{root}",
		check:  noVCSSuffix,
	},

	// Github gists
	{
		prefix: "gist.github.com/",
		re:     `^(?P<root>gist.github\.com/[A-Za-z0-9_.\-]+/?[A-Za-z0-9_.\-]+)(/[\p{L}0-9_.\-]+)*$`,
		vcs:    "git",
		repo:   "https://{root}",
		check:  noVCSSuffix,
	},

	// Bitbucket
	{
		prefix: "bitbucket.org/",
		re:     `^(?P<root>bitbucket\.org/(?P<bitname>[A-Za-z0-9_.\-]+/[A-Za-z0-9_.\-]+))(/[A-Za-z0-9_.\-]+)*$`,
		repo:   "https://{root}",
		check:  bitbucketVCS,
	},

	// IBM DevOps Services (JazzHub)
	{
		prefix: "hub.jazz.net/git/",
		re:     `^(?P<root>hub\.jazz\.net/git/[a-z0-9]+/[A-Za-z0-9_.\-]+)(/[A-Za-z0-9_.\-]+)*$`,
		vcs:    "git",
		repo:   "https://{root}",
		check:  noVCSSuffix,
	},

	// Git at Apache
	{
		prefix: "git.apache.org/",
		re:     `^(?P<root>git\.apache\.org/[a-z0-9_.\-]+\.git)(/[A-Za-z0-9_.\-]+)*$`,
		vcs:    "git",
		repo:   "https://{root}",
	},

	// Git at OpenStack
	{
		prefix: "git.openstack.org/",
		re:     `^(?P<root>git\.openstack\.org/[A-Za-z0-9_.\-]+/[A-Za-z0-9_.\-]+)(\.git)?(/[A-Za-z0-9_.\-]+)*$`,
		vcs:    "git",
		repo:   "https://{root}",
	},

	// chiselapp.com for fossil
	{
		prefix: "chiselapp.com/",
		re:     `^(?P<root>chiselapp\.com/user/[A-Za-z0-9]+/repository/[A-Za-z0-9_.\-]+)$`,
		vcs:    "fossil",
		repo:   "https://{root}",
	},

	// General syntax for any server.
	// Must be last.
	{
		re:   `^(?P<root>(?P<repo>([a-z0-9.\-]+\.)+[a-z0-9.\-]+(:[0-9]+)?(/~?[A-Za-z0-9_.\-]+)+?)\.(?P<vcs>bzr|fossil|git|hg|svn))(/~?[A-Za-z0-9_.\-]+)*$`,
		ping: true,
	},
}

// vcsPathsAfterDynamic gives additional vcsPaths entries
// to try after the dynamic HTML check.
// This gives those sites a chance to introduce <meta> tags
// as part of a graceful transition away from the hard-coded logic.
var vcsPathsAfterDynamic = []*vcsPath{
	// Launchpad. See golang.org/issue/11436.
	{
		prefix: "launchpad.net/",
		re:     `^(?P<root>launchpad\.net/((?P<project>[A-Za-z0-9_.\-]+)(?P<series>/[A-Za-z0-9_.\-]+)?|~[A-Za-z0-9_.\-]+/(\+junk|[A-Za-z0-9_.\-]+)/[A-Za-z0-9_.\-]+))(/[A-Za-z0-9_.\-]+)*$`,
		vcs:    "bzr",
		repo:   "https://{root}",
		check:  launchpadVCS,
	},
}

func init() {
	// fill in cached regexps.
	// Doing this eagerly discovers invalid regexp syntax
	// without having to run a command that needs that regexp.
	for _, srv := range vcsPaths {
		srv.regexp = regexp.MustCompile(srv.re)
	}
	for _, srv := range vcsPathsAfterDynamic {
		srv.regexp = regexp.MustCompile(srv.re)
	}
}

// noVCSSuffix checks that the repository name does not
// end in .foo for any version control system foo.
// The usual culprit is ".git".
func noVCSSuffix(ctx context.Context, match map[string]string, gitreq *gitcache.Request) error {
	repo := match["repo"]
	for _, vcs := range vcsList {
		if strings.HasSuffix(repo, "."+vcs) {
			return fmt.Errorf("invalid version control suffix in %s path", match["prefix"])
		}
	}
	return nil
}

// bitbucketVCS determines the version control system for a
// Bitbucket repository, by using the Bitbucket API.
func bitbucketVCS(ctx context.Context, match map[string]string, gitreq *gitcache.Request) error {
	if err := noVCSSuffix(ctx, match, gitreq); err != nil {
		return err
	}

	var resp struct {
		SCM string `json:"scm"`
	}
	url := expand(match, "https://api.bitbucket.org/2.0/repositories/{bitname}?fields=scm")
	data, err := Get(ctx, url)
	if err != nil {
		if httpErr, ok := err.(*HTTPError); ok && httpErr.StatusCode == 403 {
			// this may be a private repository. If so, attempt to determine which
			// VCS it uses. See issue 5375.
			root := match["root"]
			for _, vcs := range []string{"git", "hg"} {
				provider := vcsByCmd(vcs, gitreq)
				if provider != nil && provider.ping(ctx, "https", root) == nil {
					resp.SCM = vcs
					break
				}
			}
		}

		if resp.SCM == "" {
			return err
		}
	} else {
		if err := json.Unmarshal(data, &resp); err != nil {
			return fmt.Errorf("decoding %s: %v", url, err)
		}
	}

	if vcsByCmd(resp.SCM, gitreq) != nil {
		match["vcs"] = resp.SCM
		if resp.SCM == "git" {
			match["repo"] += ".git"
		}
		return nil
	}

	return fmt.Errorf("unable to detect version control system for bitbucket.org/ path")
}

// launchpadVCS solves the ambiguity for "lp.net/project/foo". In this case,
// "foo" could be a series name registered in Launchpad with its own branch,
// and it could also be the name of a directory within the main project
// branch one level up.
func launchpadVCS(ctx context.Context, match map[string]string, gitreq *gitcache.Request) error {
	if match["project"] == "" || match["series"] == "" {
		return nil
	}
	_, err := Get(ctx, expand(match, "https://code.launchpad.net/{project}{series}/.bzr/branch-format"))
	if err != nil {
		match["root"] = expand(match, "launchpad.net/{project}")
		match["repo"] = expand(match, "https://{root}")
	}
	return nil
}
