package getter

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/storage/memory"

	"regexp"

	"net/http"
	"net/url"

	"encoding/xml"
	"io"

	"crypto/tls"
	"time"

	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-git.v4/storage"
)

// downloadPackage runs the create or download command
// to make the first copy of or update a copy of the given package.
func (c *Cache) downloadPackage(p *Package, update bool, insecure bool) error {

	var root *repoRoot
	var err error

	if p.Build.SrcRoot != "" {
		// Directory exists. Look for checkout along path to src.
		root, err = c.vcsFromDir(p.Dir, p.Build.SrcRoot)
		if err != nil {
			return err
		}
	} else {
		// Analyze the import path to determine the version control system,
		// repository, and the import path for the root of the repository.
		root, err = c.repoRootForImportPath(p.ImportPath, insecure)
		if err != nil {
			return err
		}
	}
	if !isSecure(root.url) && !insecure {
		return fmt.Errorf("cannot download, %v uses insecure protocol", root.url)
	}

	if p.Build.SrcRoot == "" {
		// Package not found. Put in first directory of $GOPATH.
		list := filepath.SplitList(c.buildContext.GOPATH)
		if len(list) == 0 {
			return fmt.Errorf("cannot download, $GOPATH not set. For more details see: 'go help gopath'")
		}
		// Guard against people setting GOPATH=$GOROOT.
		if filepath.Clean(list[0]) == filepath.Clean(c.buildContext.GOROOT) {
			return fmt.Errorf("cannot download, $GOPATH must not be set to $GOROOT. For more details see: 'go help gopath'")
		}
		if _, err := c.fs.Stat(filepath.Join(list[0], "src/cmd/go/alldocs.go")); err == nil {
			return fmt.Errorf("cannot download, %s is a GOROOT, not a GOPATH. For more details see: 'go help gopath'", list[0])
		}
		p.Build.Root = list[0]
		p.Build.SrcRoot = filepath.Join(list[0], "src")
		p.Build.PkgRoot = filepath.Join(list[0], "pkg")
	}
	dir := filepath.Join(p.Build.SrcRoot, filepath.FromSlash(root.path))
	if root.dir == "" {
		root.dir = dir
	} else if root.dir != dir {
		return fmt.Errorf("path disagreement, calculated %s, expected %s", dir, root.dir)
	}

	c.repoPackages[p.ImportPath] = root

	// If we've considered this repository already, don't do it again.
	if _, ok := c.repoRoots[root.dir]; ok {
		return nil
	}
	c.repoRoots[root.dir] = root

	//if cfg.BuildV {
	//	fmt.Fprintf(os.Stderr, "%s (download)\n", rootPath)
	//}

	// TODO: remove this?
	// Check that this is an appropriate place for the repo to be checked out.
	// The target directory must either not exist or have a repo checked out already.
	//meta := filepath.Join(root.rootDir, "."+root.vcs.cmd())
	//st, err := c.fs.Stat(meta)
	//if err == nil && !st.IsDir() {
	//	return fmt.Errorf("%s exists but is not a directory", meta)
	//}
	if !root.exists {
		// Metadata directory does not exist. Prepare to checkout new copy.
		// Some version control tools require the target directory not to exist.
		// We require that too, just to avoid stepping on existing work.
		if _, err := c.fs.Stat(root.dir); err == nil {
			return fmt.Errorf("dir %s exists but repo does not", root.dir)
		}

		//_, err := c.fs.Stat(p.Build.Root)
		//gopathExisted := err == nil

		// Some version control tools require the parent of the target to exist.
		parent, _ := filepath.Split(root.dir)
		if err = c.fs.MkdirAll(parent, 0777); err != nil {
			return err
		}
		//if cfg.BuildV && !gopathExisted && p.Internal.Build.Root == cfg.BuildContext.GOPATH {
		//	fmt.Fprintf(os.Stderr, "created GOPATH=%s; see 'go help gopath'\n", p.Internal.Build.Root)
		//}

		if c.log != nil {
			fmt.Fprintf(c.log, "cloning %s\n", root.path)
		}
		if err = root.create(c.fs); err != nil {
			return err
		}
	} else {
		// Metadata directory does exist; download incremental updates.

		if c.log != nil {
			fmt.Fprintf(c.log, "pulling %s\n", root.path)
		}

		if err = root.download(); err != nil {
			return err
		}
	}

	//if cfg.BuildN {
	// Do not show tag sync in -n; it's noise more than anything,
	// and since we're not running commands, no tag will be found.
	// But avoid printing nothing.
	//	fmt.Fprintf(os.Stderr, "# cd %s; %s sync/update\n", rootDir, vcs.cmd)
	//	return nil
	//}

	// TODO: work out if we actually need this...

	// Select and sync to appropriate version of the repository.
	//tags, err := vcs.tags(rootDir)
	//if err != nil {
	//	return err
	//}
	//vers := runtime.Version()
	//if i := strings.Index(vers, " "); i >= 0 {
	//	vers = vers[:i]
	//}
	//if err := vcs.tagSync(rootDir, selectTag(vers, tags)); err != nil {
	//	return err
	//}

	return nil
}

// vcsFromDir inspects dir and its parents to determine the
// version control system and code repository to use.
// On return, root is the import path
// corresponding to the root of the repository.
func (c *Cache) vcsFromDir(dir, srcRoot string) (root *repoRoot, err error) {
	// Clean and double-check that dir is in (a subdirectory of) srcRoot.
	dir = filepath.Clean(dir)
	srcRoot = filepath.Clean(srcRoot)
	if len(dir) <= len(srcRoot) || dir[len(srcRoot)] != filepath.Separator {
		return nil, fmt.Errorf("directory %q is outside source root %q", dir, srcRoot)
	}

	origDir := dir
	for len(dir) > len(srcRoot) {
		if root, ok := c.repoRoots[dir]; ok {
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

// repoRootForImportPath analyzes importPath to determine the
// version control system, and code repository to use.
func (c *Cache) repoRootForImportPath(importPath string, insecure bool) (*repoRoot, error) {
	rr, err := repoRootFromVCSPaths(importPath, "", insecure, vcsPaths)
	if err == errUnknownSite {
		rr, err = c.repoRootForImportDynamic(importPath, insecure)
		if err != nil {
			err = fmt.Errorf("unrecognized import path %q (%v)", importPath, err)
		}
	}
	if err != nil {
		rr1, err1 := repoRootFromVCSPaths(importPath, "", insecure, vcsPathsAfterDynamic)
		if err1 == nil {
			rr = rr1
			err = nil
		}
	}

	if err == nil && strings.Contains(importPath, "...") && strings.Contains(rr.path, "...") {
		// Do not allow wildcards in the repo root.
		rr = nil
		err = fmt.Errorf("cannot expand ... in %q", importPath)
	}
	return rr, err
}

// repoRootForImportDynamic finds a *repoRoot for a custom domain that's not
// statically known by repoRootForImportPathStatic.
//
// This handles custom import paths like "name.tld/pkg/foo" or just "name.tld".
func (c *Cache) repoRootForImportDynamic(importPath string, insecure bool) (*repoRoot, error) {
	slash := strings.Index(importPath, "/")
	if slash < 0 {
		slash = len(importPath)
	}
	host := importPath[:slash]
	if !strings.Contains(host, ".") {
		return nil, errors.New("import path does not begin with hostname")
	}
	urlStr, body, err := webGetMaybeInsecure(importPath, insecure)
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
	//if cfg.BuildV {
	//	log.Printf("get %q: found meta tag %#v at %s", importPath, mmi, urlStr)
	//}
	// If the import was "uni.edu/bob/project", which said the
	// prefix was "uni.edu" and the RepoRoot was "evilroot.com",
	// make sure we don't trust Bob and check out evilroot.com to
	// "uni.edu" yet (possibly overwriting/preempting another
	// non-evil student). Instead, first verify the root and see
	// if it matches Bob's claim.
	if mmi.Prefix != importPath {
		//if cfg.BuildV {
		//	log.Printf("get %q: verifying non-authoritative meta tag", importPath)
		//}
		urlStr0 := urlStr
		var imports []metaImport
		urlStr, imports, err = c.metaImportsForPrefix(mmi.Prefix, insecure)
		if err != nil {
			return nil, err
		}
		metaImport2, err := matchGoImport(imports, importPath)
		if err != nil || mmi != metaImport2 {
			return nil, fmt.Errorf("%s and %s disagree about go-import for %s", urlStr0, urlStr, mmi.Prefix)
		}
	}

	if !strings.Contains(mmi.RepoRoot, "://") {
		return nil, fmt.Errorf("%s: invalid repo root %q; no scheme", urlStr, mmi.RepoRoot)
	}
	rr := &repoRoot{
		vcs:    vcsByCmd(mmi.VCS),
		url:    mmi.RepoRoot,
		path:   mmi.Prefix,
		custom: true,
	}
	if rr.vcs == nil {
		return nil, fmt.Errorf("%s: unknown vcs %q", urlStr, mmi.VCS)
	}
	return rr, nil
}

// metaImportsForPrefix takes a package's root import path as declared in a <meta> tag
// and returns its HTML discovery URL and the parsed metaImport lines
// found on the page.
//
// The importPath is of the form "golang.org/x/tools".
// It is an error if no imports are found.
// urlStr will still be valid if err != nil.
// The returned urlStr will be of the form "https://golang.org/x/tools?go-get=1"
func (c *Cache) metaImportsForPrefix(importPrefix string, insecure bool) (urlStr string, imports []metaImport, err error) {
	setCache := func(res fetchResult) (fetchResult, error) {
		c.fetchCacheMu.Lock()
		defer c.fetchCacheMu.Unlock()
		c.fetchCache[importPrefix] = res
		return res, nil
	}

	resi, _, _ := c.fetchGroup.Do(importPrefix, func() (resi interface{}, err error) {
		c.fetchCacheMu.Lock()
		if res, ok := c.fetchCache[importPrefix]; ok {
			c.fetchCacheMu.Unlock()
			return res, nil
		}
		c.fetchCacheMu.Unlock()

		urlStr, body, err := webGetMaybeInsecure(importPrefix, insecure)
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

// parseMetaGoImports returns meta imports from the HTML in r.
// Parsing ends at the end of the <head> section or the beginning of the <body>.
func parseMetaGoImports(r io.Reader) (imports []metaImport, err error) {
	d := xml.NewDecoder(r)
	d.CharsetReader = charsetReader
	d.Strict = false
	var t xml.Token
	for {
		t, err = d.RawToken()
		if err != nil {
			if err == io.EOF || len(imports) > 0 {
				err = nil
			}
			return
		}
		if e, ok := t.(xml.StartElement); ok && strings.EqualFold(e.Name.Local, "body") {
			return
		}
		if e, ok := t.(xml.EndElement); ok && strings.EqualFold(e.Name.Local, "head") {
			return
		}
		e, ok := t.(xml.StartElement)
		if !ok || !strings.EqualFold(e.Name.Local, "meta") {
			continue
		}
		if attrValue(e.Attr, "name") != "go-import" {
			continue
		}
		if f := strings.Fields(attrValue(e.Attr, "content")); len(f) == 3 {
			imports = append(imports, metaImport{
				Prefix:   f[0],
				VCS:      f[1],
				RepoRoot: f[2],
			})
		}
	}
}

// metaImport represents the parsed <meta name="go-import"
// content="prefix vcs reporoot" /> tags from HTML files.
type metaImport struct {
	Prefix, VCS, RepoRoot string
}

// attrValue returns the attribute value for the case-insensitive key
// `name', or the empty string if nothing is found.
func attrValue(attrs []xml.Attr, name string) string {
	for _, a := range attrs {
		if strings.EqualFold(a.Name.Local, name) {
			return a.Value
		}
	}
	return ""
}

// charsetReader returns a reader for the given charset. Currently
// it only supports UTF-8 and ASCII. Otherwise, it returns a meaningful
// error which is printed by go get, so the user can find why the package
// wasn't downloaded if the encoding is not supported. Note that, in
// order to reduce potential errors, ASCII is treated as UTF-8 (i.e. characters
// greater than 0x7f are not rejected).
func charsetReader(charset string, input io.Reader) (io.Reader, error) {
	switch strings.ToLower(charset) {
	case "ascii":
		return input, nil
	default:
		return nil, fmt.Errorf("can't decode XML document using charset %q", charset)
	}
}

// repoRootFromVCSPaths attempts to map importPath to a repoRoot
// using the mappings defined in vcsPaths.
// If scheme is non-empty, that scheme is forced.
func repoRootFromVCSPaths(importPath, scheme string, insecure bool, vcsPaths []*vcsPath) (*repoRoot, error) {
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
			if err := srv.check(match); err != nil {
				return nil, err
			}
		}
		provider := vcsByCmd(match["vcs"])
		if provider == nil {
			return nil, fmt.Errorf("unknown version control system %q", match["vcs"])
		}
		if srv.ping {
			if scheme != "" {
				match["repo"] = scheme + "://" + match["repo"]
			} else {
				for _, scheme := range provider.schemes() {
					if !insecure && !defaultSecureScheme[scheme] {
						continue
					}
					if provider.ping(scheme, match["repo"]) == nil {
						match["repo"] = scheme + "://" + match["repo"]
						break
					}
				}
			}
		}
		rr := &repoRoot{
			vcs:  provider,
			url:  match["repo"],
			path: match["root"],
		}
		return rr, nil
	}
	return nil, errUnknownSite
}

var errUnknownSite = errors.New("dynamic lookup required to find mapping")

type repoRoot struct {
	vcs vcsProvider

	// url is the repository URL, including scheme
	url string

	// dir is the dir corresponding to the root of the repository
	dir string

	// path is the import path corresponding to the root of the repository
	path string

	// custom is true for custom import paths (those defined by HTML meta tags)
	custom bool

	// exists is true once the repo has been downloaded
	exists bool

	// hash is the latest commit hash
	hash string
}

func (r *repoRoot) Hash() string {
	return r.hash
}

func (r *repoRoot) create(fs billy.Filesystem) error {
	if err := r.vcs.create(r.url, r.dir, fs); err != nil {
		return err
	}
	r.exists = true
	r.hash = r.vcs.hash()
	return nil
}

func (r *repoRoot) download() error {
	if err := r.vcs.download(); err != nil {
		return err
	}
	r.hash = r.vcs.hash()
	return nil
}

var httpPrefixRE = regexp.MustCompile(`^https?:`)

func isSecure(repo string) bool {
	u, err := url.Parse(repo)
	if err != nil {
		// If repo is not a URL, it's not secure.
		return false
	}
	return defaultSecureScheme[u.Scheme]
}

var defaultSecureScheme = map[string]bool{
	"https":   true,
	"git+ssh": true,
	"bzr+ssh": true,
	"svn+ssh": true,
	"ssh":     true,
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
		prefix: "hub.jazz.net/git",
		re:     `^(?P<root>hub.jazz.net/git/[a-z0-9]+/[A-Za-z0-9_.\-]+)(/[A-Za-z0-9_.\-]+)*$`,
		vcs:    "git",
		repo:   "https://{root}",
		check:  noVCSSuffix,
	},

	// Git at Apache
	{
		prefix: "git.apache.org",
		re:     `^(?P<root>git.apache.org/[a-z0-9_.\-]+\.git)(/[A-Za-z0-9_.\-]+)*$`,
		vcs:    "git",
		repo:   "https://{root}",
	},

	// Git at OpenStack
	{
		prefix: "git.openstack.org",
		re:     `^(?P<root>git\.openstack\.org/[A-Za-z0-9_.\-]+/[A-Za-z0-9_.\-]+)(\.git)?(/[A-Za-z0-9_.\-]+)*$`,
		vcs:    "git",
		repo:   "https://{root}",
	},

	// General syntax for any server.
	// Must be last.
	{
		re:   `^(?P<root>(?P<repo>([a-z0-9.\-]+\.)+[a-z0-9.\-]+(:[0-9]+)?(/~?[A-Za-z0-9_.\-]+)+?)\.(?P<vcs>bzr|git|hg|svn))(/~?[A-Za-z0-9_.\-]+)*$`,
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

// launchpadVCS solves the ambiguity for "lp.net/project/foo". In this case,
// "foo" could be a series name registered in Launchpad with its own branch,
// and it could also be the name of a directory within the main project
// branch one level up.
func launchpadVCS(match map[string]string) error {
	if match["project"] == "" || match["series"] == "" {
		return nil
	}
	_, err := webGet(expand(match, "https://code.launchpad.net/{project}{series}/.bzr/branch-format"))
	if err != nil {
		match["root"] = expand(match, "launchpad.net/{project}")
		match["repo"] = expand(match, "https://{root}")
	}
	return nil
}

// A vcsPath describes how to convert an import path into a
// version control system and repository name.
type vcsPath struct {
	prefix string                              // prefix this description applies to
	re     string                              // pattern for import path
	repo   string                              // repository to use (expand with match of re)
	vcs    string                              // version control system to use (expand with match of re)
	check  func(match map[string]string) error // additional checks
	ping   bool                                // ping for scheme to use to download repo

	regexp *regexp.Regexp // cached compiled form of re
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
func noVCSSuffix(match map[string]string) error {
	repo := match["repo"]
	for _, cmd := range vcsList {
		if strings.HasSuffix(repo, "."+cmd) {
			return fmt.Errorf("invalid version control suffix in %s path", match["prefix"])
		}
	}
	return nil
}

// bitbucketVCS determines the version control system for a
// Bitbucket repository, by using the Bitbucket API.
func bitbucketVCS(match map[string]string) error {
	if err := noVCSSuffix(match); err != nil {
		return err
	}

	var resp struct {
		SCM string `json:"scm"`
	}
	url := expand(match, "https://api.bitbucket.org/2.0/repositories/{bitname}?fields=scm")
	data, err := webGet(url)
	if err != nil {
		if httpErr, ok := err.(*HTTPError); ok && httpErr.StatusCode == 403 {
			// this may be a private repository. If so, attempt to determine which
			// VCS it uses. See issue 5375.
			root := match["root"]
			for _, cmd := range []string{"git", "hg"} {
				provider := vcsByCmd(cmd)
				if provider != nil && provider.ping("https", root) == nil {
					resp.SCM = cmd
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

	if vcsByCmd(resp.SCM) != nil {
		match["vcs"] = resp.SCM
		if resp.SCM == "git" {
			match["repo"] += ".git"
		}
		return nil
	}

	return fmt.Errorf("unable to detect version control system for bitbucket.org/ path")
}

// expand rewrites s to replace {k} with match[k] for each key k in match.
func expand(match map[string]string, s string) string {
	for k, v := range match {
		s = strings.Replace(s, "{"+k+"}", v, -1)
	}
	return s
}

func vcsByCmd(cmd string) vcsProvider {
	switch cmd {
	case "git":
		return new(gitProvider)
	}
	return nil
}

type vcsProvider interface {
	cmd() string
	ping(scheme, repo string) error
	schemes() []string
	create(url, dir string, fs billy.Filesystem) error
	download() error
	hash() string
}

type gitProvider struct {
	store      storage.Storer
	repo       *git.Repository
	worktree   *git.Worktree
	hashString string
}

func (g *gitProvider) hash() string {
	return g.hashString
}

func (g *gitProvider) create(url, dir string, fs billy.Filesystem) error {
	// git clone {repo} {dir}
	// git -go-internal-cd {dir} submodule update --init --recursive
	g.store = memory.NewStorage()
	dirfs, err := fs.Chroot(dir)
	if err != nil {
		return err
	}
	repo, err := git.Clone(g.store, dirfs, &git.CloneOptions{
		URL:               url,
		SingleBranch:      true,
		Depth:             1,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})
	if err != nil {
		return err
	}
	g.repo = repo

	worktree, err := g.repo.Worktree()
	if err != nil {
		return err
	}
	g.worktree = worktree

	// ... retrieves the branch pointed by HEAD
	ref, err := repo.Head()
	if err != nil {
		return err
	}

	// ... retrieves the commit history
	iter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return err
	}

	c, err := iter.Next()
	if err != nil {
		return err
	}

	g.hashString = c.Hash.String()

	return nil
}

func (g *gitProvider) download() error {
	// git pull --ff-only
	// git submodule update --init --recursive
	err := g.worktree.Pull(&git.PullOptions{
		SingleBranch:      true,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		Force:             true,
	})
	if err != nil {
		return err
	}

	// ... retrieves the branch pointed by HEAD
	ref, err := g.repo.Head()
	if err != nil {
		return err
	}

	// ... retrieves the commit history
	iter, err := g.repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return err
	}

	c, err := iter.Next()
	if err != nil {
		return err
	}

	g.hashString = c.Hash.String()

	return nil
}

func (g *gitProvider) cmd() string {
	return "git"
}

func (g *gitProvider) schemes() []string {
	return []string{"git", "https", "http", "git+ssh", "ssh"}
}

func (g *gitProvider) ping(scheme, repo string) error {
	repository, _ := git.Init(memory.NewStorage(), nil)

	// Add a new remote, with the default fetch refspec
	remote, err := repository.CreateRemote(&config.RemoteConfig{
		Name: "example",
		URLs: []string{scheme + "://" + repo},
	})
	if err != nil {
		return err
	}
	_, err = remote.List(&git.ListOptions{})
	return err
}

// vcsList lists the known version control systems
var vcsList = []string{
	"git",
}

// httpClient is the default HTTP client, but a variable so it can be
// changed by tests, without modifying http.DefaultClient.
var httpClient = http.DefaultClient

// impatientInsecureHTTPClient is used in -insecure mode,
// when we're connecting to https servers that might not be there
// or might be using self-signed certificates.
var impatientInsecureHTTPClient = &http.Client{
	Timeout: 5 * time.Second,
	Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

type HTTPError struct {
	status     string
	StatusCode int
	url        string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("%s: %s", e.url, e.status)
}

// Get returns the data from an HTTP GET request for the given URL.
func webGet(url string) ([]byte, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err := &HTTPError{status: resp.Status, StatusCode: resp.StatusCode, url: url}

		return nil, err
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", url, err)
	}
	return b, nil
}

// GetMaybeInsecure returns the body of either the importPath's
// https resource or, if unavailable and permitted by the security mode, the http resource.
func webGetMaybeInsecure(importPath string, insecure bool) (urlStr string, body io.ReadCloser, err error) {
	fetch := func(scheme string) (urlStr string, res *http.Response, err error) {
		u, err := url.Parse(scheme + "://" + importPath)
		if err != nil {
			return "", nil, err
		}
		u.RawQuery = "go-get=1"
		urlStr = u.String()
		//if cfg.BuildV {
		//	log.Printf("Fetching %s", urlStr)
		//}
		if insecure && scheme == "https" { // fail earlier
			res, err = impatientInsecureHTTPClient.Get(urlStr)
		} else {
			res, err = httpClient.Get(urlStr)
		}
		return
	}
	closeBody := func(res *http.Response) {
		if res != nil {
			res.Body.Close()
		}
	}
	urlStr, res, err := fetch("https")
	if err != nil {
		//if cfg.BuildV {
		//	log.Printf("https fetch failed: %v", err)
		//}
		if insecure {
			closeBody(res)
			urlStr, res, err = fetch("http")
		}
	}
	if err != nil {
		closeBody(res)
		return "", nil, err
	}
	// Note: accepting a non-200 OK here, so people can serve a
	// meta import in their http 404 page.
	//if cfg.BuildV {
	//	log.Printf("Parsing meta tags from %s (status code %d)", urlStr, res.StatusCode)
	//}
	return urlStr, res.Body, nil
}
