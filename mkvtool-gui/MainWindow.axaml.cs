using System;
using System.IO;
using System.Threading.Tasks;
using Avalonia.Controls;
using Avalonia.Interactivity;
using Avalonia.Markup.Xaml;
using MessageBox.Avalonia;
using MEnums = MessageBox.Avalonia.Enums;

namespace mkvtool
{
    public partial class MainWindow : Window
    {
        public MainWindow()
        {
            InitializeComponent();
        }

        private void InitializeComponent()
        {
            AvaloniaXamlLoader.Load(this);
        }

        private async void CheckFileBtn_OnClick(object? sender, RoutedEventArgs e)
        {
            string[] files = await ShowSelectFileDialog("MKV file", new string[] {"mkv"}, false);
            if (files != null)
            {
                bool[] result = mkvlib.CheckSubset(files[0], lcb);
                if (result[1])
                    await MessageBoxManager.GetMessageBoxStandardWindow("Check result",
                        "Has error.", MEnums.ButtonEnum.Ok,
                        MEnums.Icon.Error, WindowStartupLocation.CenterOwner).Show();
                else if (result[0])
                    await MessageBoxManager.GetMessageBoxStandardWindow("Check result",
                        "This mkv file are subsetted.", MEnums.ButtonEnum.Ok,
                        MEnums.Icon.Info, WindowStartupLocation.CenterOwner).Show();
                else
                    await MessageBoxManager.GetMessageBoxStandardWindow("Check result",
                        "This mkv file are not subsetted.", MEnums.ButtonEnum.Ok,
                        MEnums.Icon.Warning, WindowStartupLocation.CenterOwner).Show();
            }
        }

        private async void CheckFolderBtn_OnClick(object? sender, RoutedEventArgs e)
        {
            string dir = await new OpenFolderDialog().ShowAsync(this);
            if (!string.IsNullOrEmpty(dir))
            {
                string[] list = mkvlib.QueryFolder(dir, lcb);
                if (list != null && list.Length > 0)
                {
                    lcb("Not subsetted file list:");
                    lcb(" ----- Begin ----- ");
                    lcb(string.Join(Environment.NewLine, list));
                    lcb(" -----  End  ----- ");
                }
                else
                    lcb("All files are subseted.");
            }
        }

        void lcb(string str)
        {
            this.FindControl<TextBox>("logBox").Text += str + Environment.NewLine;
        }

        private async void TopLevel_OnOpened(object? sender, EventArgs e)
        {
            if (!mkvlib.InitInstance(lcb))
            {
                await MessageBoxManager.GetMessageBoxStandardWindow("Check result",
                    "Failed to init mkvlib.", MEnums.ButtonEnum.Ok,
                    MEnums.Icon.Error, WindowStartupLocation.CenterOwner).Show();
                this.IsEnabled = false;
            }
        }

        class SubsetArg
        {
            public static string[] Asses { get; set; }
            public static string Fonts { get; set; }
            public static string Output { get; set; }
            public static bool DirSafe { get; set; }
        }

        private async void SubsetSelectBtns_OnClick(object? sender, RoutedEventArgs e)
        {
            Button btn = (Button) sender;
            string dir;
            switch (btn.Tag.ToString())
            {
                case "asses":
                    SubsetArg.Asses = null;
                    this.FindControl<TextBlock>("sa1").Text = string.Empty;
                    string[] files = await ShowSelectFileDialog("ASS file(s)", new[] {"ass"}, true);
                    if (files != null && files.Length > 0)
                    {
                        SubsetArg.Asses = files;
                        this.FindControl<TextBlock>("sa1").Text = string.Join(Environment.NewLine, files);
                    }

                    break;
                case "fonts":
                    SubsetArg.Fonts = string.Empty;
                    this.FindControl<TextBlock>("sa2").Text = string.Empty;
                    dir = await new OpenFolderDialog().ShowAsync(this);
                    if (!string.IsNullOrEmpty(dir))
                    {
                        SubsetArg.Fonts = dir;
                        this.FindControl<TextBlock>("sa2").Text = dir;
                    }

                    break;
                case "output":
                    SubsetArg.Output = string.Empty;
                    this.FindControl<TextBlock>("sa3").Text = string.Empty;
                    dir = await new OpenFolderDialog().ShowAsync(this);
                    if (!string.IsNullOrEmpty(dir))
                    {
                        SubsetArg.Output = dir;
                        this.FindControl<TextBlock>("sa3").Text = dir;
                    }

                    break;
            }
        }

        async Task<string[]> ShowSelectFileDialog(string name, string[] ext, bool multiple)
        {
            OpenFileDialog fileDialog = new OpenFileDialog();
            fileDialog.AllowMultiple = multiple;
            FileDialogFilter filter = new FileDialogFilter();
            filter.Name = name;
            filter.Extensions.AddRange(ext);
            fileDialog.Filters.Add(filter);
            string[] files = await fileDialog.ShowAsync(this);
            return files;
        }

        private async void DoSubsetBtn_OnClick(object? sender, RoutedEventArgs e)
        {
            if (SubsetArg.Asses != null && SubsetArg.Asses.Length > 0 && !string.IsNullOrEmpty(SubsetArg.Fonts) &&
                !string.IsNullOrEmpty(SubsetArg.Output))
            {
                SubsetArg.DirSafe = this.FindControl<CheckBox>("sa4").IsChecked == true;
                if (mkvlib.ASSFontSubset(SubsetArg.Asses, SubsetArg.Fonts, SubsetArg.Output, SubsetArg.DirSafe, lcb))
                {
                    await MessageBoxManager.GetMessageBoxStandardWindow("Subset result",
                        "Subset successfully.", MEnums.ButtonEnum.Ok,
                        MEnums.Icon.Info, WindowStartupLocation.CenterOwner).Show();
                    SubsetArg.Asses = null;
                    SubsetArg.Fonts = string.Empty;
                    SubsetArg.Output = string.Empty;
                    this.FindControl<TextBlock>("sa1").Text = string.Empty;
                    this.FindControl<TextBlock>("sa2").Text = string.Empty;
                    this.FindControl<TextBlock>("sa3").Text = string.Empty;
                    this.FindControl<CheckBox>("sa4").IsCancel = true;
                }
                else
                    await MessageBoxManager.GetMessageBoxStandardWindow("Subset result",
                        "Failed to subset.", MEnums.ButtonEnum.Ok,
                        MEnums.Icon.Info, WindowStartupLocation.CenterOwner).Show();
            }
        }

        class DumpArg
        {
            public static string Path { get; set; }
            public static string Output { get; set; }
            public static bool Subset { get; set; }
            public static bool Dir { get; set; }
        }

        private async void DumpSelectBtns_OnClick(object? sender, RoutedEventArgs e)
        {
            Button btn = (Button) sender;
            string dir;
            switch (btn.Tag.ToString())
            {
                case "file":
                    DumpArg.Path = string.Empty;
                    DumpArg.Dir = false;
                    this.FindControl<TextBlock>("da1").Text = string.Empty;
                    string[] files = await ShowSelectFileDialog("MKV file",
                        new[] {"mkv"},
                        false);
                    if (files != null && files.Length > 0)
                    {
                        DumpArg.Path = files[0];
                        this.FindControl<TextBlock>("da1").Text = files[0];
                    }

                    break;
                case "folder":
                    DumpArg.Path = string.Empty;
                    this.FindControl<TextBlock>("da1").Text = string.Empty;
                    dir = await new OpenFolderDialog().ShowAsync(this);
                    if (!string.IsNullOrEmpty(dir))
                    {
                        DumpArg.Path = dir;
                        this.FindControl<TextBlock>("da1").Text = dir;
                        DumpArg.Dir = true;
                    }

                    break;
                case "output":
                    DumpArg.Output = string.Empty;
                    this.FindControl<TextBlock>("da2").Text = string.Empty;
                    dir = await new OpenFolderDialog().ShowAsync(this);
                    if (!string.IsNullOrEmpty(dir))
                    {
                        DumpArg.Output = dir;
                        this.FindControl<TextBlock>("da2").Text = dir;
                    }

                    break;
            }
        }

        private async void DoDumpBtn_OnClick(object? sender, RoutedEventArgs e)
        {
            if (!string.IsNullOrEmpty(DumpArg.Path) &&
                !string.IsNullOrEmpty(DumpArg.Output))
            {
                DumpArg.Subset = this.FindControl<CheckBox>("da3").IsChecked == true;
                bool r = !DumpArg.Dir
                    ? mkvlib.DumpMKV(DumpArg.Path, DumpArg.Output, DumpArg.Subset, lcb)
                    : mkvlib.DumpMKVs(DumpArg.Path, DumpArg.Output, DumpArg.Subset, lcb);
                if (r)
                {
                    await MessageBoxManager.GetMessageBoxStandardWindow("Dump result",
                        "Dump successfully.", MEnums.ButtonEnum.Ok,
                        MEnums.Icon.Info, WindowStartupLocation.CenterOwner).Show();
                    DumpArg.Path = string.Empty;
                    DumpArg.Output = string.Empty;
                    DumpArg.Dir = false;
                    this.FindControl<TextBlock>("da1").Text = string.Empty;
                    this.FindControl<TextBlock>("da2").Text = string.Empty;
                    this.FindControl<CheckBox>("da3").IsCancel = true;
                }
                else
                    await MessageBoxManager.GetMessageBoxStandardWindow("Dump result",
                        "Failed to dump.", MEnums.ButtonEnum.Ok,
                        MEnums.Icon.Info, WindowStartupLocation.CenterOwner).Show();
            }
        }

        class MakeArg
        {
            public static string Dir { get; set; }
            public static string Data { get; set; }
            public static string Output { get; set; }
            public static string slang { get; set; }
            public static string stitle { get; set; }
        }

        private async void MakeSelectBtns_OnClick(object? sender, RoutedEventArgs e)
        {
            Button btn = (Button) sender;
            string dir;
            switch (btn.Tag.ToString())
            {
                case "dir":
                    MakeArg.Dir = string.Empty;
                    this.FindControl<TextBlock>("ma1").Text = string.Empty;
                    dir = await new OpenFolderDialog().ShowAsync(this);
                    if (!string.IsNullOrEmpty(dir))
                    {
                        MakeArg.Dir = dir;
                        this.FindControl<TextBlock>("ma1").Text = dir;
                    }

                    break;
                case "data":
                    MakeArg.Data = string.Empty;
                    this.FindControl<TextBlock>("ma2").Text = string.Empty;
                    dir = await new OpenFolderDialog().ShowAsync(this);
                    if (!string.IsNullOrEmpty(dir))
                    {
                        MakeArg.Data = dir;
                        this.FindControl<TextBlock>("ma2").Text = dir;
                    }

                    break;
                case "output":
                    MakeArg.Output = string.Empty;
                    this.FindControl<TextBlock>("ma3").Text = string.Empty;
                    dir = await new OpenFolderDialog().ShowAsync(this);
                    if (!string.IsNullOrEmpty(dir))
                    {
                        MakeArg.Output = dir;
                        this.FindControl<TextBlock>("ma3").Text = dir;
                    }

                    break;
            }
        }

        private async void DoMakeBtn_OnClick(object? sender, RoutedEventArgs e)
        {
            if (!string.IsNullOrEmpty(MakeArg.Dir) && !string.IsNullOrEmpty(MakeArg.Data) &&
                !string.IsNullOrEmpty(MakeArg.Output))
            {
                MakeArg.slang = this.FindControl<TextBox>("ma4").Text;
                MakeArg.stitle = this.FindControl<TextBox>("ma5").Text;
                if (mkvlib.MakeMKVs(MakeArg.Dir, MakeArg.Data, MakeArg.Output, MakeArg.slang, MakeArg.stitle, lcb))
                {
                    await MessageBoxManager.GetMessageBoxStandardWindow("Make result",
                        "Make successfully.", MEnums.ButtonEnum.Ok,
                        MEnums.Icon.Info, WindowStartupLocation.CenterOwner).Show();
                    MakeArg.Dir = string.Empty;
                    MakeArg.Data = string.Empty;
                    MakeArg.Output = string.Empty;
                    MakeArg.slang = string.Empty;
                    MakeArg.stitle = string.Empty;
                    this.FindControl<TextBlock>("ma1").Text = string.Empty;
                    this.FindControl<TextBlock>("ma2").Text = string.Empty;
                    this.FindControl<TextBlock>("ma3").Text = string.Empty;
                    this.FindControl<TextBox>("ma4").Text = string.Empty;
                    this.FindControl<TextBox>("ma5").Text = string.Empty;
                }
                else
                    await MessageBoxManager.GetMessageBoxStandardWindow("Make result",
                        "Failed to make.", MEnums.ButtonEnum.Ok,
                        MEnums.Icon.Info, WindowStartupLocation.CenterOwner).Show();
            }
        }

        class CreateArg
        {
            public static string vDir { get; set; }
            public static string sDir { get; set; }
            public static string fDir { get; set; }
            public static string oDir { get; set; }
            public static string slang { get; set; }
            public static string stitle { get; set; }
            public static bool clean { get; set; }
        }

        private async void CreateSelectBtns_OnClick(object? sender, RoutedEventArgs e)
        {
            Button btn = (Button) sender;
            string dir;
            switch (btn.Tag.ToString())
            {
                case "v":
                    CreateArg.vDir = string.Empty;
                    this.FindControl<TextBlock>("ca1").Text = string.Empty;
                    dir = await new OpenFolderDialog().ShowAsync(this);
                    if (!string.IsNullOrEmpty(dir))
                    {
                        CreateArg.vDir = dir;
                        this.FindControl<TextBlock>("ca1").Text = dir;
                    }

                    break;
                case "s":
                    CreateArg.sDir = string.Empty;
                    this.FindControl<TextBlock>("ca2").Text = string.Empty;
                    dir = await new OpenFolderDialog().ShowAsync(this);
                    if (!string.IsNullOrEmpty(dir))
                    {
                        CreateArg.sDir = dir;
                        this.FindControl<TextBlock>("ca2").Text = dir;
                    }

                    break;
                case "f":
                    CreateArg.fDir = string.Empty;
                    this.FindControl<TextBlock>("ca3").Text = string.Empty;
                    dir = await new OpenFolderDialog().ShowAsync(this);
                    if (!string.IsNullOrEmpty(dir))
                    {
                        CreateArg.fDir = dir;
                        this.FindControl<TextBlock>("ca3").Text = dir;
                    }

                    break;
                case "o":
                    CreateArg.oDir = string.Empty;
                    this.FindControl<TextBlock>("ca4").Text = string.Empty;
                    dir = await new OpenFolderDialog().ShowAsync(this);
                    if (!string.IsNullOrEmpty(dir))
                    {
                        CreateArg.oDir = dir;
                        this.FindControl<TextBlock>("ca4").Text = dir;
                    }

                    break;
            }
        }

        private async void DoCreatetn_OnClick(object? sender, RoutedEventArgs e)
        {
            if (!string.IsNullOrEmpty(CreateArg.vDir) && !string.IsNullOrEmpty(CreateArg.sDir) &&
                !string.IsNullOrEmpty(CreateArg.fDir) && !string.IsNullOrEmpty(CreateArg.oDir))
            {
                CreateArg.slang = this.FindControl<TextBox>("ca5").Text;
                CreateArg.stitle = this.FindControl<TextBox>("ca6").Text;
                CreateArg.clean = this.FindControl<CheckBox>("ca7").IsChecked == true;
                if (mkvlib.CreateMKVs(CreateArg.vDir, CreateArg.sDir, CreateArg.fDir, string.Empty, CreateArg.oDir,
                    CreateArg.slang, CreateArg.stitle, CreateArg.clean, lcb))
                {
                    await MessageBoxManager.GetMessageBoxStandardWindow("Create result",
                        "Create successfully.", MEnums.ButtonEnum.Ok,
                        MEnums.Icon.Info, WindowStartupLocation.CenterOwner).Show();
                    CreateArg.vDir = string.Empty;
                    CreateArg.sDir = string.Empty;
                    CreateArg.fDir = string.Empty;
                    CreateArg.oDir = string.Empty;
                    CreateArg.clean = false;
                    this.FindControl<TextBlock>("ca1").Text = string.Empty;
                    this.FindControl<TextBlock>("ca2").Text = string.Empty;
                    this.FindControl<TextBlock>("ca3").Text = string.Empty;
                    this.FindControl<TextBlock>("ca4").Text = string.Empty;
                    this.FindControl<TextBox>("ca5").Text = string.Empty;
                    this.FindControl<TextBox>("ca6").Text = string.Empty;
                    this.FindControl<CheckBox>("ca7").IsChecked = false;
                }
                else
                    await MessageBoxManager.GetMessageBoxStandardWindow("Create result",
                        "Failed to create.", MEnums.ButtonEnum.Ok,
                        MEnums.Icon.Info, WindowStartupLocation.CenterOwner).Show();
            }
        }
    }
}