```ssh
#vim8的安装
yum install ncurses-devel
wget https://github.com/vim/vim/archive/master.zip
unzip master.zip
cd vim-master
cd src/
./configure --enable-gui=auto --enable-gtk2-check --with-x   --with-features=huge  --prefix=/usr/local  --enable-pythoninterp=yes --with-python-config-dir=/usr/lib64/python3.4/config-3.4m
make
sudo make install
```

#install zsh
```
yum install zsh

#install oh-my-zsh
sh -c "$(curl -fsSL https://raw.githubusercontent.com/robbyrussell/oh-my-zsh/master/tools/install.sh)"

#zsh
cd ~/.oh-my-zsh/custom/plugins/
git clone https://github.com/zsh-users/zsh-syntax-highlighting
git clone https://github.com/zsh-users/zsh-autosuggestions
git clone https://github.com/zsh-users/zsh-completions
git clone https://github.com/zsh-users/zsh-history-substring-search
```
#vim
## intall vim plugin plug
curl -fLo ~/.vim/autoload/plug.vim --create-dirs https://raw.githubusercontent.com/junegunn/vim-plug/master/plug.vim

## intall vim monokai themes
cd .vim &&  mkdir colors && cd colors && git clone https://github.com/sickill/vim-monokai && cp vim-monokai/colors/vim-monokai.vim . && rm -rf vim-monokai


vim ~/.zshrc
plugins=(
  git
  zsh-autosuggestions
  zsh-syntax-highlighting
  zsh-completions
  zsh-history-substring-search
)

```
